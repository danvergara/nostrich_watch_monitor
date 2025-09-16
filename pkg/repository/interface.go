package repository

import (
	"context"

	"github.com/danvergara/nostrich_watch_monitor/pkg/domain"
)

type ListOption struct {
	Limit  *int
	Offset *int
	URLs   []string
}

type RelayRepository interface {
	Create(ctx context.Context, relayInfo domain.Relay) error
	List(ctx context.Context, opts *ListOption) ([]domain.Relay, error)
	FindByURL(ctx context.Context, url string) (domain.Relay, error)
	Update(ctx context.Context, relayInfo domain.Relay) error
	SaveHealthCheck(ctx context.Context, status domain.HealthCheck) error
}
