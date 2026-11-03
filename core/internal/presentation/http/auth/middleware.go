// Package auth checks HTTP Basic credentials on every /v1 route. The 401 carries
// no WWW-Authenticate header: the browser must not open its own dialog.
package auth

import (
	"net/http"

	"fluffy-giggle/core/internal/domain/errs"
	httprender "fluffy-giggle/core/internal/platform/http/render"
	"fluffy-giggle/core/internal/usecase/authenticate"
)

func Middleware(uc *authenticate.Usecase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			login, password, ok := r.BasicAuth()
			if !ok || !uc.Execute(login, password) {
				httprender.Unauthorized(w, errs.ErrUnauthenticated)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
