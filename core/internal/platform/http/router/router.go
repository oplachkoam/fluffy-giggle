package httprouter

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/fx"

	httpmiddleware "fluffy-giggle/core/internal/platform/http/middleware"
)

type Registrar interface {
	RegisterRoutes(router chi.Router)
}

type Params struct {
	fx.In

	Registrars []Registrar `group:"routes"`
	// Middlewares wrap everything under /v1 and nothing else: /health and the SPA
	// stay open. Authentication lives here.
	Middlewares []httpmiddleware.Middleware `group:"v1-middlewares"`
	Fallback    http.Handler                `                       name:"fallback" optional:"true"`
}

func New(params Params) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer, httpmiddleware.Logging())
	router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	router.Route("/v1", func(router chi.Router) {
		router.Use(params.Middlewares...)
		for _, registrar := range params.Registrars {
			registrar.RegisterRoutes(router)
		}
	})
	if params.Fallback != nil {
		router.Handle("/*", params.Fallback)
	}

	return router
}
