package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// GetFS returns the filesystem rooted at the compiled web dist directory.
func GetFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}

// SPAHandler wraps the embedded filesystem to support client-side routing.
func SPAHandler() (http.Handler, error) {
	sub, err := GetFS()
	if err != nil {
		return nil, err
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Try opening file
		f, err := sub.Open(path)
		if err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback to index.html for SPA routes
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	}), nil
}
