package routes

import (
	"github.com/go-chi/chi/v5"

	httprouter "fluffy-giggle/core/internal/platform/http/router"
	"fluffy-giggle/core/internal/presentation/http/handler/createrun"
	"fluffy-giggle/core/internal/presentation/http/handler/events"
	"fluffy-giggle/core/internal/presentation/http/handler/getbenchmark"
	"fluffy-giggle/core/internal/presentation/http/handler/getrun"
	"fluffy-giggle/core/internal/presentation/http/handler/getrunimage"
	"fluffy-giggle/core/internal/presentation/http/handler/listbenchmarks"
	"fluffy-giggle/core/internal/presentation/http/handler/listcases"
	"fluffy-giggle/core/internal/presentation/http/handler/listruns"
	"fluffy-giggle/core/internal/presentation/http/handler/me"
	"fluffy-giggle/core/internal/presentation/http/handler/resolvereview"
	"fluffy-giggle/core/internal/presentation/http/handler/startbenchmark"
)

type Routes struct {
	me             *me.Handler
	events         *events.Handler
	createRun      *createrun.Handler
	listRuns       *listruns.Handler
	getRun         *getrun.Handler
	getRunImage    *getrunimage.Handler
	resolveReview  *resolvereview.Handler
	listCases      *listcases.Handler
	startBenchmark *startbenchmark.Handler
	getBenchmark   *getbenchmark.Handler
	listBenchmarks *listbenchmarks.Handler
}

var _ httprouter.Registrar = (*Routes)(nil)

func New(
	me *me.Handler,
	events *events.Handler,
	createRun *createrun.Handler,
	listRuns *listruns.Handler,
	getRun *getrun.Handler,
	getRunImage *getrunimage.Handler,
	resolveReview *resolvereview.Handler,
	listCases *listcases.Handler,
	startBenchmark *startbenchmark.Handler,
	getBenchmark *getbenchmark.Handler,
	listBenchmarks *listbenchmarks.Handler,
) *Routes {
	return &Routes{
		me: me, events: events, createRun: createRun, listRuns: listRuns, getRun: getRun,
		getRunImage: getRunImage, resolveReview: resolveReview, listCases: listCases,
		startBenchmark: startBenchmark, getBenchmark: getBenchmark, listBenchmarks: listBenchmarks,
	}
}

func (r *Routes) RegisterRoutes(router chi.Router) {
	router.Get("/auth/me", r.me.Handle)
	router.Get("/events", r.events.Handle)
	router.Post("/runs", r.createRun.Handle)
	router.Get("/runs", r.listRuns.Handle)
	router.Get("/runs/{id}", r.getRun.Handle)
	router.Get("/runs/{id}/image", r.getRunImage.Handle)
	router.Post("/runs/{id}/review", r.resolveReview.Handle)
	router.Get("/cases", r.listCases.Handle)
	router.Post("/benchmarks", r.startBenchmark.Handle)
	router.Get("/benchmarks", r.listBenchmarks.Handle)
	router.Get("/benchmarks/{id}", r.getBenchmark.Handle)
}
