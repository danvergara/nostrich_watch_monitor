package server

import (
	"net/http"

	"github.com/danvergara/nostrich_watch_monitor/internal/config"
	"github.com/danvergara/nostrich_watch_monitor/pkg/domain"
	"github.com/danvergara/nostrich_watch_monitor/web/views"
)

func GetMockRelays() []domain.RelayDisplayData {
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

	mockRelays := []domain.RelayDisplayData{
		{
			URL:              "wss://relay.damus.io",
			Name:             "Damus Relay",
			Description:      "High-performance relay optimized for mobile clients with excellent uptime and fast response times.",
			UptimePercent:    99.2,
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
			Description:      "Premium paid relay with advanced spam filtering and guaranteed message delivery.",
			UptimePercent:    98.7,
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
			Description:      "Community-driven relay focused on social interactions and content discovery.",
			UptimePercent:    97.1,
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
			Description:      "Web of Trust relay with curated content and verified user network.",
			UptimePercent:    96.8,
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
			Description:      "Fun and fast relay with a focus on memes, jokes, and light-hearted content.",
			UptimePercent:    94.2,
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
		views.Dashboard(GetMockRelays()).Render(r.Context(), w)
	}
}
