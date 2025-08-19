/*
Copyright © 2025 Daniel Vergara daniel.omar.vergara@gmail.com
*/
package cmd

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/spf13/cobra"

	"github.com/danvergara/nostrich_watch_monitor/web"
	"github.com/danvergara/nostrich_watch_monitor/web/views"
)

//go:generate tailwindcss -i ./web/static/css/input.css -o ./web/static/css/styles.css --minify

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "HTTP server for the dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		dashboardHandler := http.NewServeMux()

		staticFs, err := fs.Sub(web.StaticFiles, "static")
		if err != nil {
			return err
		}

		dashboardHandler.Handle(
			"/static/",
			http.StripPrefix("/static/", http.FileServer(http.FS(staticFs))),
		)
		dashboard := views.Dashboard()
		dashboardHandler.Handle("/", templ.Handler(dashboard))

		if err := http.ListenAndServe(":8080", dashboardHandler); err != nil {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
