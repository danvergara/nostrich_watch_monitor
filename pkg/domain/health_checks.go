package domain

import (
	"database/sql/driver"
	"time"

	"github.com/lib/pq"
)

type IntArray []int

func (a *IntArray) Scan(value any) error {
	return pq.Array(a).Scan(value)
}

func (a IntArray) Value() (driver.Value, error) {
	return pq.Array(a).Value()
}

// HealthCheck is a struct that maps the health_checks on the PostgreSQL database.
// It represents the online status of the given relay.
type HealthCheck struct {
	RelayURL         string    `db:"relay_url"`
	CreatedAt        time.Time `db:"created_at"`
	WebsocketSuccess bool      `db:"websocket_success"`
	WebsocketError   *string   `db:"websocket_error"`
	Nip11Success     *bool     `db:"nip11_success"`
	Nip11Error       *string   `db:"nip11_error"`
	RTTOpen          *int      `db:"rtt_open"`
	RTTRead          *int      `db:"rtt_read"`
	RTTWrite         *int      `db:"rtt_write"`
	RTTNIP11         *int      `db:"rtt_nip11"`
}
