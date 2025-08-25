package domain

import (
	"time"

	"github.com/lib/pq"
)

// Relay is a struct that maps the relays table on the PostgreSQL database.
// It represents the NIP-11 relay information with database tags for sqlx.
type Relay struct {
	URL string `db:"url"`

	Name        *string `db:"name"`
	Description *string `db:"description"`
	PubKey      *string `db:"pubkey"`
	Contact     *string `db:"contact"`

	Icon           *string `db:"icon"`
	Banner         *string `db:"banner"`
	PrivacyPolicy  *string `db:"privacy_policy"`
	TermsOfService *string `db:"terms_of_service"`

	Software *string `db:"software"`
	Version  *string `db:"version"`

	SupportedNIPs  pq.Int64Array  `db:"supported_nips"`
	RelayCountries pq.StringArray `db:"relay_countries"`
	LanguageTags   pq.StringArray `db:"language_tags"`

	Tags          pq.StringArray `db:"tags"`
	PostingPolicy *string        `db:"posting_policy"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	*HealthCheck `db:"health_checks"`
}
