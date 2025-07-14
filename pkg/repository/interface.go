package repository

import (
	"context"

	"github.com/danvergara/nostrich_watch_monitor/pkg/domain"
)

type RelayRepository interface {
	List(ctx context.Context) ([]domain.Relay, error)
}
