package server

import (
	"io/fs"
	"net/http"

	"github.com/a-h/templ"

	"github.com/danvergara/nostrich_watch_monitor/internal/config"
	"github.com/danvergara/nostrich_watch_monitor/web/views"
)

func addRoutes(mux *http.ServeMux, cfg *config.Config, fs fs.FS) {
	mux.Handle(
		"/static/",
		http.StripPrefix("/static/", http.FileServer(http.FS(fs))),
	)

	dashboard := views.Dashboard()
	mux.Handle("/", templ.Handler(dashboard))
}
