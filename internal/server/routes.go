package server

import (
	"io/fs"
	"net/http"

	"github.com/danvergara/nostrich_watch_monitor/internal/config"
)

func addRoutes(mux *http.ServeMux, cfg *config.Config, fs fs.FS) {
	mux.Handle(
		"/static/",
		http.StripPrefix("/static/", http.FileServer(http.FS(fs))),
	)

	mux.HandleFunc("/", dashboarIndexHandler(cfg))
}
