package healthcheck

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip11"
)

// convertAnyToInt utility function to convert an slice of any to an slice of integers.
// Moslty used for supported NIPS from the NIP 11 response.
func convertAnyToInt(input []any) ([]int, error) {
	var result []int

	for i, value := range input {
		switch v := value.(type) {
		case int:
			result = append(result, v)
		case float64:
			result = append(result, int(v))
		default:
			return nil, fmt.Errorf("element at index %d not supported - real type is %T", i, value)
		}
	}

	return result, nil
}

// Helper functions for nullable database fields.
func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullBool(b bool) *bool {
	return &b
}

func nullInt(i *int) int {
	if i == nil {
		return 0
	}

	return *i
}

// addSupportedNIPs helper function to add the supported NIPs from the NIP11 response to the 30166 events as tags.
func addSupportedNIPs(tags nostr.Tags, supportedNIPs []int) nostr.Tags {
	for _, n := range supportedNIPs {
		tags = append(tags, nostr.Tag{"N", strconv.Itoa(n)})
	}

	return tags
}

// addLimitations helper function to add the 30116 event if the relays requires payment or authentication.
func addLimitations(tags nostr.Tags, limitations *nip11.RelayLimitationDocument) nostr.Tags {
	if limitations != nil {
		if limitations.PaymentRequired {
			tags = append(tags, nostr.Tag{"R", "payment"})
		} else {
			tags = append(tags, nostr.Tag{"R", "!payment"})
		}

		if limitations.AuthRequired {
			tags = append(tags, nostr.Tag{"R", "auth"})
		} else {
			tags = append(tags, nostr.Tag{"R", "!auth"})
		}
	}

	return tags
}

// addTopics helper function to add the topics of the relay from the NIP11 response to the 30166 event.
func addTopics(tags nostr.Tags, topics []string) nostr.Tags {
	for _, t := range topics {
		tags = append(tags, nostr.Tag{"t", t})
	}

	return tags
}

// addLanguages helper function to add the "l" tags to the 30166 event based on the "language_tags" field from the NIP11 response.
func addLanguages(tags nostr.Tags, languageTags []string) nostr.Tags {
	for _, langTag := range languageTags {
		standard := getLanguageTagStandard(langTag)
		tags = append(tags, nostr.Tag{"l", langTag, standard})
	}
	return tags
}

// getLanguageTagStandard helper function to determine the language tag standard
func getLanguageTagStandard(tag string) string {
	// Global marker is part of BCP-47.
	if tag == "*" {
		return "BCP-47"
	}

	parts := strings.Split(tag, "-")

	// Simple 2-letter codes are ISO-639-1.
	if len(tag) == 2 {
		return "ISO-639-1"
	}

	// 3-letter codes are ISO-639-2 or ISO-639-3.
	if len(tag) == 3 {
		return "ISO-639-2"
	}

	// Anything with hyphens (regions, scripts) is BCP-47.
	if len(parts) > 1 {
		return "BCP-47"
	}

	// Default to BCP-47 as it's the most comprehensive.
	return "BCP-47"
}
