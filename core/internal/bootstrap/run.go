// Package bootstrap builds the application graph: config, platform, application
// code, HTTP server.
package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"fluffy-giggle/core/internal/anonymizer"
	"fluffy-giggle/core/internal/config"
	"fluffy-giggle/core/internal/domain/port"
	"fluffy-giggle/core/internal/fixture"
	"fluffy-giggle/core/internal/llm"
	sqlitedb "fluffy-giggle/core/internal/platform/database/sqlite"
	httpmiddleware "fluffy-giggle/core/internal/platform/http/middleware"
	httprouter "fluffy-giggle/core/internal/platform/http/router"
	httpserver "fluffy-giggle/core/internal/platform/http/server"
	"fluffy-giggle/core/internal/platform/observability/logger"
	"fluffy-giggle/core/internal/platform/validation"
	httpauth "fluffy-giggle/core/internal/presentation/http/auth"
	createrunhandler "fluffy-giggle/core/internal/presentation/http/handler/createrun"
	eventshandler "fluffy-giggle/core/internal/presentation/http/handler/events"
	getbenchmarkhandler "fluffy-giggle/core/internal/presentation/http/handler/getbenchmark"
	getrunhandler "fluffy-giggle/core/internal/presentation/http/handler/getrun"
	getrunimagehandler "fluffy-giggle/core/internal/presentation/http/handler/getrunimage"
	listbenchmarkshandler "fluffy-giggle/core/internal/presentation/http/handler/listbenchmarks"
	listcaseshandler "fluffy-giggle/core/internal/presentation/http/handler/listcases"
	listrunshandler "fluffy-giggle/core/internal/presentation/http/handler/listruns"
	"fluffy-giggle/core/internal/presentation/http/handler/me"
	resolvereviewhandler "fluffy-giggle/core/internal/presentation/http/handler/resolvereview"
	startbenchmarkhandler "fluffy-giggle/core/internal/presentation/http/handler/startbenchmark"
	"fluffy-giggle/core/internal/presentation/http/routes"
	"fluffy-giggle/core/internal/presentation/http/ui"
	"fluffy-giggle/core/internal/storage/memory/eventbus"
	benchmarkrepo "fluffy-giggle/core/internal/storage/sqlite/benchmark"
	runrepo "fluffy-giggle/core/internal/storage/sqlite/run"
	steprepo "fluffy-giggle/core/internal/storage/sqlite/step"
	"fluffy-giggle/core/internal/usecase/authenticate"
	"fluffy-giggle/core/internal/usecase/createrun"
	"fluffy-giggle/core/internal/usecase/getbenchmark"
	"fluffy-giggle/core/internal/usecase/getrun"
	"fluffy-giggle/core/internal/usecase/getrunimage"
	"fluffy-giggle/core/internal/usecase/listbenchmarks"
	"fluffy-giggle/core/internal/usecase/listcases"
	"fluffy-giggle/core/internal/usecase/listruns"
	"fluffy-giggle/core/internal/usecase/resolvereview"
	"fluffy-giggle/core/internal/usecase/runpipeline"
	"fluffy-giggle/core/internal/usecase/startbenchmark"
	"fluffy-giggle/core/internal/usecase/streamevents"
	"fluffy-giggle/core/migrations"
)

func Run() {
	fx.New(Options()).Run()
}

func Options() fx.Option {
	return fx.Options(
		fx.Provide(config.Load),
		Module,
		fx.Invoke(func(*http.Server) {}),
	)
}

// Module is the application without config loading: tests supply *config.Config.
var Module = fx.Options(
	fx.WithLogger(func() fxevent.Logger {
		return &fxevent.SlogLogger{Logger: slog.Default()}
	}),
	fx.Provide(splitConfig),
	fx.Invoke(logger.Setup),
	fx.Provide(
		validation.New,
		openDatabase,
		fx.Annotate(runrepo.New, fx.As(new(port.RunRepository))),
		fx.Annotate(steprepo.New, fx.As(new(port.StepRepository))),
		fx.Annotate(benchmarkrepo.New, fx.As(new(port.BenchmarkRepository))),
		fx.Annotate(eventbus.New, fx.As(new(port.EventBus))),
		fx.Annotate(anonymizer.New, fx.As(new(port.Detector))),
		fx.Annotate(fixture.New, fx.As(new(port.CaseCatalog))),
		newModels,
		newAuthenticate,
		newPipeline,
		newWorker,
		fx.Annotate(newWorkerQueue, fx.As(new(port.RunQueue))),
		createrun.New, resolvereview.New, getrun.New, getrunimage.New, listruns.New,
		listcases.New, startbenchmark.New, getbenchmark.New, listbenchmarks.New, streamevents.New,
		me.New, eventshandler.New, createrunhandler.New, listrunshandler.New, getrunhandler.New,
		getrunimagehandler.New, resolvereviewhandler.New, listcaseshandler.New,
		startbenchmarkhandler.New, getbenchmarkhandler.New, listbenchmarkshandler.New,
		fx.Annotate(routes.New, fx.As(new(httprouter.Registrar)), fx.ResultTags(`group:"routes"`)),
		fx.Annotate(newAuthMiddleware, fx.ResultTags(`group:"v1-middlewares"`)),
		fx.Annotate(ui.New, fx.ResultTags(`name:"fallback"`)),
		httprouter.New,
		httpserver.New,
	),
)

type configs struct {
	fx.Out

	Logger   logger.Config
	Server   httpserver.Config
	SQLite   sqlitedb.Config
	Auth     config.Auth
	LLM      config.LLM
	Pipeline config.Pipeline
}

func splitConfig(cfg *config.Config) configs {
	return configs{
		Logger: cfg.Logger, Server: cfg.HTTPServer, SQLite: cfg.SQLite,
		Auth: cfg.Auth, LLM: cfg.LLM, Pipeline: cfg.Pipeline,
	}
}

func openDatabase(lc fx.Lifecycle, cfg sqlitedb.Config) (*sql.DB, error) {
	db, err := sqlitedb.Open(cfg)
	if err != nil {
		return nil, err
	}
	if err := sqlitedb.Migrate(context.Background(), db, migrations.Files); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	lc.Append(fx.Hook{OnStop: func(context.Context) error { return db.Close() }})
	return db, nil
}

func newModels(cfg config.LLM) port.Models {
	client := func(m config.Model) port.Model {
		return llm.New(llm.Config{
			BaseURL:         m.BaseURL,
			APIKey:          m.APIKey,
			Model:           m.Model,
			Timeout:         m.Timeout,
			MaxRetries:      m.MaxRetries,
			DisableThinking: m.DisableThinking, Temperature: m.Temperature,
		})
	}
	return port.Models{Local: client(cfg.Local), Vision: client(cfg.Vision), Cloud: client(cfg.Cloud)}
}

func newAuthenticate(cfg config.Auth) *authenticate.Usecase {
	return authenticate.New(cfg.Login, cfg.Password)
}

func newAuthMiddleware(uc *authenticate.Usecase) httpmiddleware.Middleware {
	return httpauth.Middleware(uc)
}

func newPipeline(
	runs port.RunRepository, steps port.StepRepository, benchmarks port.BenchmarkRepository,
	models port.Models, detector port.Detector, events port.EventBus, cfg config.Pipeline,
) *runpipeline.Usecase {
	return runpipeline.New(runs, steps, benchmarks, models, detector, events,
		runpipeline.Config{Gate2MaxPasses: cfg.Gate2MaxPasses})
}

func newWorker(lc fx.Lifecycle, uc *runpipeline.Usecase, cfg config.Pipeline) *runpipeline.Worker {
	worker := runpipeline.NewWorker(uc, cfg.Concurrency)
	lc.Append(fx.Hook{OnStart: worker.Start, OnStop: worker.Stop})
	return worker
}

func newWorkerQueue(worker *runpipeline.Worker) *runpipeline.Worker {
	return worker
}

// Ensure the validator is referenced even when only usecases take it.
var _ = validator.New
