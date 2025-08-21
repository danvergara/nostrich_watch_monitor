package domain

// Relay is a struct that maps the relays table on the PostgreSQL database.
// It represents the NIP-11 relay information with database tags for sqlx.
type Relay struct {
	URL            string   `db:"url"`
	Name           string   `db:"name"`
	Description    string   `db:"description"`
	PubKey         string   `db:"pubkey"`
	Contact        string   `db:"contact"`
	SupportedNIPs  IntArray `db:"supported_nips"`
	Software       string   `db:"software"`
	Version        string   `db:"version"`
	Icon           string   `db:"icon"`
	Banner         string   `db:"banner"`
	PrivacyPolicy  string   `db:"privacy_policy"`
	TermsOfService string   `db:"terms_of_service"`
	RelayCountries []string `db:"relay_countries"`
	LanguageTags   []string `db:"language_tags"`
	Tags           []string `db:"tags"`
	PostingPolicy  string   `db:"posting_policy"`
}
