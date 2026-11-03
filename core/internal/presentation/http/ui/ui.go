// Package ui serves the built SPA from the binary: a file from the bundle when it
// exists, index.html for everything else so client-side routes work.
package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var dist embed.FS

const indexFile = "index.html"

func New() (http.Handler, error) {
	files, err := fs.Sub(dist, "dist")
	if err != nil {
		return nil, err
	}
	return NewHandler(files), nil
}

func NewHandler(files fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name != "" && name != indexFile && isFile(files, name) {
			if strings.HasPrefix(name, "assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			http.ServeFileFS(w, r, files, name)
			return
		}
		if !isFile(files, indexFile) {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		r.URL.Path = "/"
		http.ServeFileFS(w, r, files, indexFile)
	})
}

func isFile(files fs.FS, name string) bool {
	for _, segment := range strings.Split(name, "/") {
		if strings.HasPrefix(segment, ".") {
			return false
		}
	}
	info, err := fs.Stat(files, name)
	return err == nil && !info.IsDir()
}
