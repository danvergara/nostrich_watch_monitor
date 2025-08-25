package server

import (
	"net/http"

	"github.com/danvergara/nostrich_watch_monitor/internal/config"
	"github.com/danvergara/nostrich_watch_monitor/pkg/presentation"
	"github.com/danvergara/nostrich_watch_monitor/web/views"
)

func GetMockRelays() []presentation.RelayTableViewModel {
	rttOpen45 := 45
	rttOpen32 := 32
	rttOpen78 := 78
	rttOpen56 := 56

	rttNip1125 := 25
	rttNip1140 := 40
	rttNip1135 := 35
	rttNip1150 := 50

	nip11Success := true
	nip11Failed := false

	mockRelays := []presentation.RelayTableViewModel{
		{
			URL:              "wss://relay.damus.io",
			Name:             "Damus Relay",
			UptimePercent:    0, // Hidden from MVP
			Classification:   "Public",
			RTTOpen:          &rttOpen45,
			RTTNIP11:         &rttNip1125,
			IsOnline:         true,
			LastCheckTime:    "2 min ago",
			WebsocketSuccess: true,
			NIP11Success:     &nip11Success,
		},
		{
			URL:              "wss://nostr.wine",
			Name:             "Nostr Wine",
			UptimePercent:    0, // Hidden from MVP
			Classification:   "Paid",
			RTTOpen:          &rttOpen32,
			RTTNIP11:         &rttNip1140,
			IsOnline:         true,
			LastCheckTime:    "2 min ago",
			WebsocketSuccess: true,
			NIP11Success:     &nip11Success,
		},
		{
			URL:              "wss://relay.snort.social",
			Name:             "Snort Social",
			UptimePercent:    0, // Hidden from MVP
			Classification:   "Public",
			RTTOpen:          &rttOpen78,
			RTTNIP11:         &rttNip1135,
			IsOnline:         true,
			LastCheckTime:    "2 min ago",
			WebsocketSuccess: true,
			NIP11Success:     &nip11Success,
		},
		{
			URL:              "wss://relay.current.fyi",
			Name:             "Current",
			UptimePercent:    0, // Hidden from MVP
			Classification:   "WoT",
			RTTOpen:          &rttOpen56,
			RTTNIP11:         &rttNip1150,
			IsOnline:         true,
			LastCheckTime:    "2 min ago",
			WebsocketSuccess: true,
			NIP11Success:     &nip11Success,
		},
		{
			URL:              "wss://nos.lol",
			Name:             "nos.lol",
			UptimePercent:    0, // Hidden from MVP
			Classification:   "Public",
			RTTOpen:          nil, // Offline, no RTT data
			RTTNIP11:         nil,
			IsOnline:         false,
			LastCheckTime:    "2 min ago",
			WebsocketSuccess: false,
			NIP11Success:     &nip11Failed,
		},
	}

	return mockRelays
}

func dashboarIndexHandler(_ *config.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = views.Dashboard(GetMockRelays()).Render(r.Context(), w)
	}
}

func GetMockRelayDetail(url string) presentation.RelayDetailViewModel {
	rttOpen45 := 45
	rttRead120 := 120
	rttWrite80 := 80
	rttNip1125 := 25

	avgRttOpen50 := 50
	avgRttRead130 := 130
	avgRttWrite85 := 85
	avgRttNip1130 := 30

	return presentation.RelayDetailViewModel{
		// Basic Info
		URL:         url,
		Name:        "Damus Relay",
		Description: "High-performance relay optimized for mobile clients with excellent uptime and fast response times. Supports all major NIPs and provides reliable message delivery.",
		Contact:     "admin@damus.io",
		PubKey:      "32e1827635450ebb3c5a7d12c1f8e7b2b514439ac10a67eef3d9fd9c5c68e245",

		// Current Status
		IsOnline:        true,
		LastCheckTime:   "2 min ago",
		CurrentRTTOpen:  &rttOpen45,
		CurrentRTTRead:  &rttRead120,
		CurrentRTTWrite: &rttWrite80,
		CurrentRTTNIP11: &rttNip1125,

		// Aggregated Health Data
		UptimePercent: 0, // Hidden from MVP
		AvgRTTOpen:    &avgRttOpen50,
		AvgRTTRead:    &avgRttRead130,
		AvgRTTWrite:   &avgRttWrite85,
		AvgRTTNIP11:   &avgRttNip1130,
		TotalChecks:   1440, // 24 hours * 60 checks per hour
		FailedChecks:  12,

		// Technical Info
		Software:      "strfry",
		Version:       "0.9.6",
		SupportedNIPs: []int{1, 2, 4, 9, 11, 12, 15, 16, 20, 22, 28, 33, 40, 42, 45, 50, 51, 65},
		Countries:     []string{"US"},
		LanguageTags:  []string{"en", "en-US"},
		Tags:          []string{"public", "mobile-optimized", "high-performance"},

		// Policy URLs
		PrivacyPolicy:  "https://damus.io/privacy",
		TermsOfService: "https://damus.io/terms",
		PostingPolicy:  "https://damus.io/posting-policy",
		Icon:           "https://damus.io/img/logo.png",
		Banner:         "https://damus.io/img/banner.jpg",

		// Classification
		Classification: "Public",
	}
}

func relayDetailHandler(_ *config.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		relayURL := r.URL.Query().Get("url")
		if relayURL == "" {
			http.Error(w, "URL parameter required", http.StatusBadRequest)
			return
		}

		// Get mock relay detail data
		relayData := GetMockRelayDetail(relayURL)

		// Render the RelayDetail template (uncomment after running templ generate)
		_ = views.RelayDetail(relayData).Render(r.Context(), w)
	}
}
