package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/IsKenKenYa/Commory/backend/internal/auth"
	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	imports, err := s.store.ListImports(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	identities, err := s.service.Identities(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	events, err := s.service.Timeline(r.Context(), buildSearchParams(r, userID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	totalEvents := 0
	for _, item := range imports {
		totalEvents += item.EventCount
	}
	lastActivity := ""
	if len(events) > 0 {
		lastActivity = events[0].Timestamp
	}
	writeJSON(w, http.StatusOK, "ok", map[string]any{
		"importCount":   len(imports),
		"identityCount": len(identities),
		"eventCount":    totalEvents,
		"lastActivity":  lastActivity,
		"recentImports": takeImports(imports, 5),
		"recentEvents":  takeEvents(events, 8),
	})
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	items, err := s.service.Events(r.Context(), buildSearchParams(r, auth.UserIDFromContext(r.Context())))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "ok", items)
}

func (s *Server) handleEvent(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/events/")
	item, err := s.service.Event(r.Context(), auth.UserIDFromContext(r.Context()), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "ok", item)
}

func (s *Server) handleTimeline(w http.ResponseWriter, r *http.Request) {
	items, err := s.service.Timeline(r.Context(), buildSearchParams(r, auth.UserIDFromContext(r.Context())))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "ok", items)
}

func (s *Server) handleIdentities(w http.ResponseWriter, r *http.Request) {
	items, err := s.service.Identities(r.Context(), auth.UserIDFromContext(r.Context()))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "ok", items)
}

func (s *Server) handleIdentity(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/identities/")
	item, err := s.service.Identity(r.Context(), auth.UserIDFromContext(r.Context()), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "ok", item)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	items, err := s.service.Search(r.Context(), buildSearchParams(r, auth.UserIDFromContext(r.Context())))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "ok", items)
}

func (s *Server) handleThread(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/threads/")
	items, err := s.service.Thread(r.Context(), auth.UserIDFromContext(r.Context()), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "ok", items)
}

func buildSearchParams(r *http.Request, userID string) msglayer.SearchParams {
	q := r.URL.Query()
	params := msglayer.SearchParams{
		UserID:      userID,
		Keyword:     q.Get("q"),
		ContactID:   q.Get("contact"),
		Type:        q.Get("type"),
		Participant: q.Get("participant"),
		From:        q.Get("from"),
		To:          q.Get("to"),
		Limit:       100,
	}
	if limit := q.Get("limit"); limit != "" {
		fmt.Sscanf(limit, "%d", &params.Limit)
	}
	if offset := q.Get("offset"); offset != "" {
		fmt.Sscanf(offset, "%d", &params.Offset)
	}
	return params
}

func takeImports(items []storage.ImportSummary, n int) []storage.ImportSummary {
	if len(items) <= n {
		return items
	}
	return items[:n]
}

func takeEvents(items []msglayer.TimelineItem, n int) []msglayer.TimelineItem {
	if len(items) <= n {
		return items
	}
	return items[:n]
}
