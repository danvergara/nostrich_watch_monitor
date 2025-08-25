package presentation

// RelayTableViewModel represents relay data optimized for dashboard table display
type RelayTableViewModel struct {
	URL              string
	Name             string
	IsOnline         bool
	UptimePercent    float64
	Classification   string // "Public", "Paid", "WoT", "Private"
	RTTOpen          *int   // WebSocket connection time (ms)
	RTTNIP11         *int   // NIP-11 fetch time (ms)
	LastCheckTime    string // When the last check cycle ran
	WebsocketSuccess bool
	NIP11Success     *bool
}

// RelayDetailViewModel represents comprehensive relay data for detail pages
// Note: This is defined but not used yet - reserved for future detail page implementation
type RelayDetailViewModel struct {
	// Basic Info (from domain.Relay)
	URL         string
	Name        string
	Description string
	Contact     string
	PubKey      string

	// Current Status (from latest health_check)
	IsOnline        bool
	LastCheckTime   string
	CurrentRTTOpen  *int
	CurrentRTTRead  *int
	CurrentRTTWrite *int
	CurrentRTTNIP11 *int

	// Aggregated Health Data
	UptimePercent float64
	AvgRTTOpen    *int
	AvgRTTRead    *int
	AvgRTTWrite   *int
	AvgRTTNIP11   *int
	TotalChecks   int
	FailedChecks  int

	// Technical Info (from domain.Relay)
	Software      string
	Version       string
	SupportedNIPs []int
	Countries     []string
	LanguageTags  []string
	Tags          []string

	// Policy URLs (from domain.Relay)
	PrivacyPolicy  string
	TermsOfService string
	PostingPolicy  string
	Icon           string
	Banner         string

	// Classification
	Classification string // derived from tags/countries
}
