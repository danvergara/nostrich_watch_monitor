package handlers

import (
	"net/http"
	"strconv"

	"github.com/danvergara/nostrich_watch_monitor/pkg/domain"
	"github.com/danvergara/nostrich_watch_monitor/pkg/presentation"
	"github.com/danvergara/nostrich_watch_monitor/pkg/services"
	"github.com/danvergara/nostrich_watch_monitor/web/views"
	"github.com/danvergara/nostrich_watch_monitor/web/views/components"
)

type RelaysHandler struct {
	service services.RelayService
}

func NewRelaysHandler(service services.RelayService) *RelaysHandler {
	return &RelaysHandler{service}
}

func (rh *RelaysHandler) HandleRelayIndex(w http.ResponseWriter, r *http.Request) {
	limit := 10
	offset := 0
	filters := &services.RelayFilters{
		Limit:  &limit,
		Offset: &offset,
	}

	relays, err := rh.service.GetRelays(r.Context(), filters)
	if err != nil {
		// Show dashboard with empty table - template will show error state via EmptyState component
		if err := views.Dashboard(ToRelayTableViewModels([]domain.Relay{})).Render(r.Context(), w); err != nil {
			// If even empty dashboard fails, try once more (template rendering rarely fails twice)
			_ = views.Dashboard(ToRelayTableViewModels([]domain.Relay{})).Render(r.Context(), w)
		}
		return
	}

	if err := views.Dashboard(ToRelayTableViewModels(relays)).Render(r.Context(), w); err != nil {
		// Same approach - show empty dashboard instead of breaking the page
		_ = views.Dashboard(ToRelayTableViewModels([]domain.Relay{})).Render(r.Context(), w)
	}
}

func (rh *RelaysHandler) HandleRelayDetail(w http.ResponseWriter, r *http.Request) {
	relayURL := r.URL.Query().Get("url")
	if relayURL == "" {
		errorRelay := createErrorRelayViewModel(
			"wss://empty-url.io",
			"URL parameter required",
		)
		_ = views.RelayDetail(errorRelay).Render(r.Context(), w)
		return
	}

	relay, err := rh.service.GetRelayByURL(r.Context(), relayURL)
	if err != nil {
		// Create error state relay view instead of breaking the page
		errorRelay := createErrorRelayViewModel(
			relayURL,
			"Relay not found or temporarily unavailable",
		)
		_ = views.RelayDetail(errorRelay).Render(r.Context(), w)
		return
	}

	if err := views.RelayDetail(ToRelayDetailViewModel(relay)).Render(r.Context(), w); err != nil {
		// Same approach - show error state instead of breaking
		errorRelay := createErrorRelayViewModel(relayURL, "Error loading relay details")
		_ = views.RelayDetail(errorRelay).Render(r.Context(), w)
	}
}

func (rh *RelaysHandler) HandleRelayRows(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // default
	offset := 0 // default

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	filters := &services.RelayFilters{
		Limit:  &limit,
		Offset: &offset,
	}

	relays, err := rh.service.GetRelays(r.Context(), filters)
	if err != nil {
		// Return error row instead of breaking the page
		if err := components.ErrorRow("Failed to load more relays. Please try again.").Render(r.Context(), w); err != nil {
			// Final fallback
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Return only the table rows, not the full page
	// Use graceful error handling instead of http.Error
	if err := components.RelayTableRows(ToRelayTableViewModels(relays), offset+limit).Render(r.Context(), w); err != nil {
		// Return error row that preserves table structure
		if err := components.ErrorRow("Failed to render relay data. Please try again.").Render(r.Context(), w); err != nil {
			// Final fallback
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
}

// Helper function to create error state relay
func createErrorRelayViewModel(url, errorMessage string) presentation.RelayDetailViewModel {
	return presentation.RelayDetailViewModel{
		URL:            url,
		Name:           "Relay Unavailable",
		Description:    errorMessage,
		IsOnline:       false,
		Classification: "Unknown",
		LastCheckTime:  "Never",
		// All other fields will be zero values, template should handle gracefully
	}
}
