package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/IsKenKenYa/Commory/backend/internal/auth"
	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

type sessionResponse struct {
	ID         string `json:"id"`
	DeviceName string `json:"device_name"`
	DeviceType string `json:"device_type"`
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
	CreatedAt  string `json:"created_at"`
	LastSeenAt string `json:"last_seen_at"`
	Current    bool   `json:"current"`
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	currentSessionID := auth.SessionIDFromContext(r.Context())
	switch r.Method {
	case http.MethodGet:
		sessions, err := s.store.ListSessionsByUser(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		result := make([]sessionResponse, 0, len(sessions))
		for _, session := range sessions {
			result = append(result, toSessionResponse(session, currentSessionID))
		}
		writeJSON(w, http.StatusOK, "ok", result)
	case http.MethodDelete:
		if currentSessionID == "" {
			writeError(w, http.StatusBadRequest, "current session unavailable")
			return
		}
		sessions, err := s.store.ListSessionsByUser(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		for _, session := range sessions {
			if session.ID == currentSessionID {
				continue
			}
			if session.RefreshTokenID != "" {
				_ = s.store.RevokeRefreshTokenByID(r.Context(), session.RefreshTokenID)
			}
		}
		if err := s.store.RevokeOtherSessions(r.Context(), userID, currentSessionID); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		_ = s.auth.WriteAuditLog(r.Context(), userID, "sessions_revoke_other", r.Header.Get("X-Forwarded-For"), r.Header.Get("User-Agent"), "")
		writeJSON(w, http.StatusOK, "ok", map[string]any{"success": true})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	sessionID := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "session id required")
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	session, err := s.store.GetSession(r.Context(), sessionID)
	if err != nil || session.UserID != userID {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if session.RefreshTokenID != "" {
		_ = s.store.RevokeRefreshTokenByID(r.Context(), session.RefreshTokenID)
	}
	if err := s.store.RevokeSession(r.Context(), sessionID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sessionID == auth.SessionIDFromContext(r.Context()) {
		clearRefreshCookie(w, s.cfg.TLS)
	}
	_ = s.auth.WriteAuditLog(r.Context(), userID, "session_revoke", r.Header.Get("X-Forwarded-For"), r.Header.Get("User-Agent"), fmt.Sprintf(`{"sessionId":%q}`, sessionID))
	writeJSON(w, http.StatusOK, "ok", map[string]any{"success": true})
}

func toSessionResponse(session storage.SessionRecord, currentSessionID string) sessionResponse {
	return sessionResponse{
		ID:         session.ID,
		DeviceName: session.DeviceName,
		DeviceType: session.DeviceType,
		IPAddress:  session.IPAddress,
		UserAgent:  session.UserAgent,
		CreatedAt:  session.CreatedAt.Format(timeFormatRFC3339),
		LastSeenAt: session.LastSeenAt.Format(timeFormatRFC3339),
		Current:    session.ID == currentSessionID,
	}
}

const timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"
