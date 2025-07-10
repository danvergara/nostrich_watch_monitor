package domain

// Relay is a struct that maps the relays table on the PostgreSQL database.
type Relay struct {
	URL string `db:"url"`
}
