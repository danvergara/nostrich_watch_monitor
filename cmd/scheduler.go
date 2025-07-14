/*
Copyright © 2025 Daniel Vergara daniel.omar.vergara@gmail.com
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cobra"

	"github.com/danvergara/nostrich_watch_monitor/pkg/database"
	"github.com/danvergara/nostrich_watch_monitor/pkg/repository/postgres"
)

var (
	dbHost string
	dbPort string
	dbUser string
	dbPass string
	dbName string
)

// schedulerCmd represents the scheduler command
var schedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "scheduler command automatically enqueues tasks to monitor relays.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		dbConfig := database.Config{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPass,
			DBName:   dbName,
		}

		db, err := database.NewPostgresDB(dbConfig)
		if err != nil {
			return err
		}
		defer db.Close()

		relayRepo := postgres.NewRelayRepository(db)
		relays, err := relayRepo.List(ctx)
		if err != nil {
			return err
		}

		// Create a context for graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s, err := gocron.NewScheduler()
		if err != nil {
			return err
		}

		// Ensure shutdown when done
		defer func() {
			if err := s.Shutdown(); err != nil {
				fmt.Printf("error shutting down scheduler: %v\n", err)
			}
		}()

		// Create a job for each relay.
		jobs := make([]gocron.Job, 0, len(relays))

		for _, r := range relays {
			// Create a job for this specific relay
			job, err := s.NewJob(
				gocron.DurationJob(15*time.Minute),
				gocron.NewTask(func() {
					fmt.Printf("healthcheck on relay: %s\n", r.URL)
				}),
				gocron.WithContext(ctx),
				gocron.WithName(fmt.Sprintf("health check on relay: %s", r.URL)),
				gocron.WithTags("health-check", "monitoring", r.URL),
			)
			if err != nil {
				fmt.Printf("error scheduling job: %v\n", err)
				continue
			}
			jobs = append(jobs, job)
		}

		// Start the scheduler.
		s.Start()
		fmt.Println("scheduler started. Task will run every 15 minutes.")

		// Show next run times.
		fmt.Println("\nNext run times:")
		for _, job := range jobs {
			if nextRun, err := job.NextRun(); err == nil {
				fmt.Printf("  %s: %s\n", job.Name(), nextRun.Format("2006-01-02 15:04:05"))
			}
		}

		// Set up graceful shutdown.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		fmt.Println("\nPress Ctrl+C to stop...")
		<-c

		fmt.Println("\nshutting down health check scheduler...")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(schedulerCmd)
}
