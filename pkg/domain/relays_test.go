package domain

import (
	"testing"
	"time"
)

func TestRelay_IsOnline(t *testing.T) {
	now := time.Now()
	healthCheckInterval := 1 * time.Minute
	truePtr := true
	falsePtr := false

	testCases := []struct {
		name    string
		relay   *Relay
		want    bool
	}{
		{
			name: "Online - recent and successful health check",
			relay: &Relay{
				HealthCheck: &HealthCheck{
					WebsocketSuccess: &truePtr,
					CreatedAt:        &now,
				},
			},
			want: true,
		},
		{
			name: "Offline - no health check",
			relay: &Relay{},
			want: false,
		},
		{
			name: "Offline - WebsocketSuccess is nil",
			relay: &Relay{
				HealthCheck: &HealthCheck{
					WebsocketSuccess: nil,
					CreatedAt:        &now,
				},
			},
			want: false,
		},
		{
			name: "Offline - WebsocketSuccess is false",
			relay: &Relay{
				HealthCheck: &HealthCheck{
					WebsocketSuccess: &falsePtr,
					CreatedAt:        &now,
				},
			},
			want: false,
		},
		{
			name: "Offline - CreatedAt is nil",
			relay: &Relay{
				HealthCheck: &HealthCheck{
					WebsocketSuccess: &truePtr,
					CreatedAt:        nil,
				},
			},
			want: false,
		},
		{
			name: "Offline - health check is too old",
			relay: &Relay{
				HealthCheck: &HealthCheck{
					WebsocketSuccess: &truePtr,
					CreatedAt:        func() *time.Time { t := now.Add(-3 * healthCheckInterval); return &t }(),
				},
			},
			want: false,
		},
        {
			name: "Online - health check is almost too old",
			relay: &Relay{
				HealthCheck: &HealthCheck{
					WebsocketSuccess: &truePtr,
					CreatedAt:        func() *time.Time { t := now.Add(-1 * time.Minute * 59 / 60 * 2); return &t }(),
				},
			},
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.relay.IsOnline(healthCheckInterval)
			if got != tc.want {
				t.Errorf("Relay.IsOnline() = %v, want %v", got, tc.want)
			}
		})
	}
}
