package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"go.uber.org/fx"
)

type Config struct {
	Host         string        `env:"HOST"          env-required:"true" validate:"required"`
	Port         int           `env:"PORT"          env-required:"true" validate:"min=1,max=65535"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT"  env-required:"true"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" env-required:"true"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT"  env-required:"true"`
}

func New(lc fx.Lifecycle, cfg Config, handler http.Handler) *http.Server {
	server := &http.Server{
		Addr:         net.JoinHostPort(cfg.Host, fmt.Sprint(cfg.Port)),
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", server.Addr)
			if err != nil {
				return fmt.Errorf("listen %s: %w", server.Addr, err)
			}
			slog.Info("http server started", slog.String("addr", server.Addr))
			go func() {
				if err := server.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
					slog.Error("http server stopped", slog.String("error", err.Error()))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})

	return server
}
