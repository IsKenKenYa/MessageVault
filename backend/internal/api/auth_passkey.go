package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/IsKenKenYa/Commory/backend/internal/auth"
	"github.com/go-webauthn/webauthn/protocol"
)

func (s *Server) handlePasskeyRegisterBegin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if s.passkey == nil {
		writeError(w, http.StatusServiceUnavailable, "passkey not configured")
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	cc, challengeID, err := s.passkey.BeginRegistration(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "ok", map[string]any{
		"options":     cc,
		"challengeId": challengeID,
	})
}

func (s *Server) handlePasskeyRegisterFinish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if s.passkey == nil {
		writeError(w, http.StatusServiceUnavailable, "passkey not configured")
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	var req struct {
		ChallengeID string `json:"challengeId"`
		Response    any    `json:"response"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	respJSON, _ := json.Marshal(req.Response)
	parsed, err := protocol.ParseCredentialCreationResponseBody(strings.NewReader(string(respJSON)))
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid response: %v", err))
		return
	}
	if err := s.passkey.FinishRegistration(r.Context(), userID, req.ChallengeID, parsed); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	_ = s.auth.WriteAuditLog(r.Context(), userID, "passkey_register", r.Header.Get("X-Forwarded-For"), r.Header.Get("User-Agent"), "")
	writeJSON(w, http.StatusOK, "ok", map[string]any{"success": true})
}

func (s *Server) handlePasskeyLoginBegin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if s.passkey == nil {
		writeError(w, http.StatusServiceUnavailable, "passkey not configured")
		return
	}
	assertion, challengeID, err := s.passkey.BeginLogin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "ok", map[string]any{
		"options":     assertion,
		"challengeId": challengeID,
	})
}

func (s *Server) handlePasskeyLoginFinish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if s.passkey == nil {
		writeError(w, http.StatusServiceUnavailable, "passkey not configured")
		return
	}
	var req struct {
		ChallengeID string `json:"challengeId"`
		Response    any    `json:"response"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	respJSON, _ := json.Marshal(req.Response)
	parsed, err := protocol.ParseCredentialRequestResponseBody(strings.NewReader(string(respJSON)))
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid response: %v", err))
		return
	}
	userID, err := s.passkey.FinishLogin(r.Context(), req.ChallengeID, parsed)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	user, err := s.auth.UserInfo(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	userRecord, err := s.store.GetUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	pair, _, err := s.auth.IssueTokenPairForUser(
		r.Context(),
		userRecord,
		r.Header.Get("X-Commory-Device"),
		r.Header.Get("X-Forwarded-For"),
		r.Header.Get("User-Agent"),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = s.auth.WriteAuditLog(r.Context(), userID, "login", r.Header.Get("X-Forwarded-For"), r.Header.Get("User-Agent"), `{"method":"passkey"}`)
	setRefreshCookie(w, pair.RefreshToken, s.cfg.TLS)
	writeJSON(w, http.StatusOK, "ok", map[string]any{
		"user":         user,
		"token":        pair.AccessToken,
		"refreshToken": pair.RefreshToken,
	})
}

func (s *Server) handlePasskeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if s.passkey == nil {
		writeError(w, http.StatusServiceUnavailable, "passkey not configured")
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	passkeys, err := s.passkey.ListPasskeys(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "ok", passkeys)
}

func (s *Server) handlePasskeyDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if s.passkey == nil {
		writeError(w, http.StatusServiceUnavailable, "passkey not configured")
		return
	}
	passkeyID := strings.TrimPrefix(r.URL.Path, "/api/auth/passkey/")
	if passkeyID == "" {
		writeError(w, http.StatusBadRequest, "passkey id required")
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	if err := s.passkey.DeletePasskey(r.Context(), userID, passkeyID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = s.auth.WriteAuditLog(r.Context(), userID, "passkey_delete", r.Header.Get("X-Forwarded-For"), r.Header.Get("User-Agent"), fmt.Sprintf(`{"passkeyId":%q}`, passkeyID))
	writeJSON(w, http.StatusOK, "ok", map[string]any{"success": true})
}
