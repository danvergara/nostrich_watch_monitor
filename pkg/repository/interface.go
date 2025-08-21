package repository

import (
	"context"

	"github.com/danvergara/nostrich_watch_monitor/pkg/domain"
)

type ListOption struct {
	Limit  *int
	Offset *int
}

type RelayRepository interface {
	List(ctx context.Context, opts *ListOption) ([]domain.Relay, error)
	Update(ctx context.Context, relayInfo domain.Relay) error
	SaveHealthCheck(ctx context.Context, status domain.HealthCheck) error
}
