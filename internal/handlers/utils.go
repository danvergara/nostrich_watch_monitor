package handlers

import (
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/danvergara/nostrich_watch_monitor/pkg/domain"
	"github.com/danvergara/nostrich_watch_monitor/pkg/presentation"
)

// ToRelayDetailViewModel converts a domain.Relay to presentation.RelayDetailViewModel
func ToRelayDetailViewModel(relay domain.Relay) presentation.RelayDetailViewModel {
	vm := presentation.RelayDetailViewModel{
		// Basic Info
		URL: relay.URL,
		Name: func() string {
			if relay.Name != nil && *relay.Name != "" {
				return *relay.Name
			}
			// Default to URL if name is empty
			return relay.URL
		}(),
		Description: safeString(relay.Description),
		Contact:     safeString(relay.Contact),
		PubKey:      safeString(relay.PubKey),

		// Technical Info
		Software:      safeString(relay.Software),
		Version:       safeString(relay.Version),
		SupportedNIPs: int64ArrayToIntSlice(relay.SupportedNIPs),
		Countries:     stringArrayToSlice(relay.RelayCountries),
		LanguageTags:  stringArrayToSlice(relay.LanguageTags),
		Tags:          stringArrayToSlice(relay.Tags),

		// Policy URLs
		PrivacyPolicy:  safeString(relay.PrivacyPolicy),
		TermsOfService: safeString(relay.TermsOfService),
		PostingPolicy:  safeString(relay.PostingPolicy),
		Icon:           safeString(relay.Icon),
		Banner:         safeString(relay.Banner),

		// Classification (derived from tags)
		Classification: deriveClassification(relay.Tags),
	}

	// Current Status (from embedded health check)
	if relay.HealthCheck != nil {
		vm.IsOnline = safeBool(relay.HealthCheck.WebsocketSuccess)
		if relay.HealthCheck.CreatedAt != nil {
			vm.LastCheckTime = FormatRelativeTime(*relay.HealthCheck.CreatedAt)
		}
		vm.CurrentRTTOpen = relay.HealthCheck.RTTOpen
		vm.CurrentRTTRead = relay.HealthCheck.RTTRead
		vm.CurrentRTTWrite = relay.HealthCheck.RTTWrite
		vm.CurrentRTTNIP11 = relay.HealthCheck.RTTNIP11
	}

	return vm
}

// FormatRelativeTime converts a time to a human-readable relative format
func FormatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case diff < 365*24*time.Hour:
		return t.Format("Jan 2")
	default:
		return t.Format("Jan 2, 2006")
	}
}

// safeString safely converts a nullable string pointer to a string
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// safeBool safely converts a nullable bool pointer to a bool
func safeBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// int64ArrayToIntSlice converts a PostgreSQL int64 array to a Go int slice
func int64ArrayToIntSlice(arr pq.Int64Array) []int {
	result := make([]int, len(arr))
	for i, v := range arr {
		result[i] = int(v)
	}
	return result
}

// stringArrayToSlice converts a PostgreSQL string array to a Go string slice
func stringArrayToSlice(arr pq.StringArray) []string {
	return []string(arr)
}

// deriveClassification determines relay classification based on tags
func deriveClassification(tags pq.StringArray) string {
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	if tagSet["paid"] {
		return "Paid"
	}
	if tagSet["wot"] {
		return "WoT"
	}
	if tagSet["private"] {
		return "Private"
	}
	return "Public"
}

// ToRelayTableViewModels converts a slice of domain.Relay to a slice of presentation.RelayTableViewModel
func ToRelayTableViewModels(relays []domain.Relay) []presentation.RelayTableViewModel {
	viewModels := make([]presentation.RelayTableViewModel, len(relays))

	for i, relay := range relays {
		viewModels[i] = ToRelayTableViewModel(relay)
	}

	return viewModels
}

// ToRelayTableViewModel converts a single domain.Relay to presentation.RelayTableViewModel
func ToRelayTableViewModel(relay domain.Relay) presentation.RelayTableViewModel {
	vm := presentation.RelayTableViewModel{
		URL: relay.URL,
		Name: func() string {
			if relay.Name != nil && *relay.Name != "" {
				return *relay.Name
			}
			// Default to URL if name is empty
			return relay.URL
		}(),
		Classification: deriveClassification(relay.Tags),
	}

	// Current Status (from embedded health check)
	if relay.HealthCheck != nil {
		vm.IsOnline = safeBool(relay.HealthCheck.WebsocketSuccess)
		vm.WebsocketSuccess = safeBool(relay.HealthCheck.WebsocketSuccess)
		vm.NIP11Success = relay.HealthCheck.Nip11Success
		vm.RTTOpen = relay.HealthCheck.RTTOpen
		vm.RTTNIP11 = relay.HealthCheck.RTTNIP11

		if relay.HealthCheck.CreatedAt != nil {
			vm.LastCheckTime = FormatRelativeTime(*relay.HealthCheck.CreatedAt)
		}
	}

	return vm
}
