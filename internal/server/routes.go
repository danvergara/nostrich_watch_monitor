package server

import (
	"io/fs"
	"net/http"

	"github.com/danvergara/nostrich_watch_monitor/internal/config"
	"github.com/danvergara/nostrich_watch_monitor/internal/handlers"
)

func addRoutes(mux *http.ServeMux, cfg *config.Config, fs fs.FS, handler handlers.RelaysHandler) {
	mux.Handle(
		"/static/",
		http.StripPrefix("/static/", http.FileServer(http.FS(fs))),
	)

	mux.HandleFunc("/", handler.HandleRelayIndex)
	mux.HandleFunc("/relay", handler.HandleRelayDetail)
}
