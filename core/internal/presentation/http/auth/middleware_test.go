package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"fluffy-giggle/core/internal/usecase/authenticate"
)

func TestMiddleware(t *testing.T) {
	t.Parallel()

	handler := Middleware(authenticate.New("operator", "correct horse"))(
		http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) },
		),
	)
	cases := []struct {
		name            string
		login, password string
		withAuth        bool
		want            int
	}{
		{"valid", "operator", "correct horse", true, http.StatusNoContent},
		{"wrong password", "operator", "wrong", true, http.StatusUnauthorized},
		{"wrong login", "admin", "correct horse", true, http.StatusUnauthorized},
		{"missing header", "", "", false, http.StatusUnauthorized},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/runs", nil)
			if tt.withAuth {
				req.SetBasicAuth(tt.login, tt.password)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			assert.Equal(t, tt.want, rec.Code)
			assert.Empty(t, rec.Header().Get("WWW-Authenticate"))
		})
	}
}
