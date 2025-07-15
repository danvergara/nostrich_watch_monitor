package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/danvergara/nostrich_watch_monitor/pkg/domain"
	"github.com/danvergara/nostrich_watch_monitor/pkg/repository"
)

type relayRepository struct {
	db *sqlx.DB
}

func NewRelayRepository(db *sqlx.DB) repository.RelayRepository {
	return &relayRepository{db: db}
}

// List returns a list of relays of interest from the databased to be monitored.
func (r *relayRepository) List(ctx context.Context) ([]domain.Relay, error) {
	var relays = []domain.Relay{}

	// Select automatically scans the query results into a slice of structs.
	err := r.db.SelectContext(ctx, &relays, "SELECT url FROM relays")
	if err != nil {
		return nil, err
	}

	return relays, nil
}
