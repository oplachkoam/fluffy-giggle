package ui

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	files := fstest.MapFS{
		"index.html":    {Data: []byte("<html>app</html>")},
		"assets/app.js": {Data: []byte("js")},
	}
	handler := NewHandler(files)

	for _, tt := range []struct {
		path, body, cache string
	}{
		{"/", "<html>app</html>", "no-cache"},
		{"/runs/123", "<html>app</html>", "no-cache"},
		{"/assets/app.js", "js", "public, max-age=31536000, immutable"},
	} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.path, nil))
		assert.Equal(t, http.StatusOK, rec.Code, tt.path)
		assert.Equal(t, tt.body, rec.Body.String(), tt.path)
		assert.Equal(t, tt.cache, rec.Header().Get("Cache-Control"), tt.path)
	}
}
