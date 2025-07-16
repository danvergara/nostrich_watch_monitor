package task

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
)

const (
	TypeHealthCheck = "relay:healthcheck"
)

// Payload for any task related to health checks on Nostr relays.
type RelayHealthCheckTaskPayload struct {
	// URL of the relay
	RelayURL string
}

func NewRelayHealthCheckTask(relayURL string) (*asynq.Task, error) {
	payload, err := json.Marshal(RelayHealthCheckTaskPayload{RelayURL: relayURL})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeHealthCheck, payload), nil
}

func HandleRelayHealthCheckTask(ctx context.Context, t *asynq.Task) error {
	var r RelayHealthCheckTaskPayload
	if err := json.Unmarshal(t.Payload(), &r); err != nil {
		return err
	}
	log.Printf("[*] Health checking %s Nostr relay", r.RelayURL)
	return nil
}
