package handlers

import (
	"net/http"

	"github.com/danvergara/nostrich_watch_monitor/pkg/domain"
	"github.com/danvergara/nostrich_watch_monitor/pkg/services"
	"github.com/danvergara/nostrich_watch_monitor/web/views"
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
		_ = views.Dashboard(ToRelayTableViewModels([]domain.Relay{})).Render(r.Context(), w)
		return
	}

	if err := views.Dashboard(ToRelayTableViewModels(relays)).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering the relays index", http.StatusInternalServerError)
		return
	}
}

func (rh *RelaysHandler) HandleRelayDetail(w http.ResponseWriter, r *http.Request) {
	relayURL := r.URL.Query().Get("url")
	if relayURL == "" {
		http.Error(w, "URL parameter required", http.StatusBadRequest)
		return
	}

	relay, err := rh.service.GetRelayByURL(r.Context(), relayURL)
	if err != nil {
		http.Error(w, "could not find relay with url", http.StatusNotFound)
		return
	}

	if err = views.RelayDetail(ToRelayDetailViewModel(relay)).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering the relay detail page", http.StatusInternalServerError)
		return
	}
}
