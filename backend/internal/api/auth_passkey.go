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
		writeError(w, http.StatusServiceUnavailable, auth.ErrOperationFailed.Error())
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	cc, challengeID, err := s.passkey.BeginRegistration(r.Context(), userID)
	if err != nil {
		logAuthInternalError("passkey register begin", err)
		writePublicAuthError(w, err)
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
		writeError(w, http.StatusServiceUnavailable, auth.ErrOperationFailed.Error())
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	var req struct {
		ChallengeID string `json:"challengeId"`
		Response    any    `json:"response"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writePublicAuthError(w, auth.ErrInvalidRequest)
		return
	}
	respJSON, _ := json.Marshal(req.Response)
	parsed, err := protocol.ParseCredentialCreationResponseBody(strings.NewReader(string(respJSON)))
	if err != nil {
		writePublicAuthError(w, auth.ErrPasskeyInvalidResponse)
		return
	}
	if err := s.passkey.FinishRegistration(r.Context(), userID, req.ChallengeID, parsed); err != nil {
		if auth.PublicErrorCode(err) == auth.ErrOperationFailed.Error() {
			logAuthInternalError("passkey register finish", err)
		}
		writePublicAuthError(w, err)
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
		writeError(w, http.StatusServiceUnavailable, auth.ErrOperationFailed.Error())
		return
	}
	assertion, challengeID, err := s.passkey.BeginLogin(r.Context())
	if err != nil {
		logAuthInternalError("passkey login begin", err)
		writePublicAuthError(w, err)
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
		writeError(w, http.StatusServiceUnavailable, auth.ErrOperationFailed.Error())
		return
	}
	var req struct {
		ChallengeID string `json:"challengeId"`
		Response    any    `json:"response"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writePublicAuthError(w, auth.ErrInvalidRequest)
		return
	}
	respJSON, _ := json.Marshal(req.Response)
	parsed, err := protocol.ParseCredentialRequestResponseBody(strings.NewReader(string(respJSON)))
	if err != nil {
		writePublicAuthError(w, auth.ErrPasskeyInvalidResponse)
		return
	}
	userID, err := s.passkey.FinishLogin(r.Context(), req.ChallengeID, parsed)
	if err != nil {
		if auth.PublicErrorCode(err) == auth.ErrOperationFailed.Error() {
			logAuthInternalError("passkey login finish", err)
		}
		writePublicAuthError(w, err)
		return
	}
	user, err := s.auth.UserInfo(r.Context(), userID)
	if err != nil {
		logAuthInternalError("passkey user info", err)
		writePublicAuthError(w, auth.ErrOperationFailed)
		return
	}
	userRecord, err := s.store.GetUser(r.Context(), userID)
	if err != nil {
		logAuthInternalError("passkey get user", err)
		writePublicAuthError(w, auth.ErrOperationFailed)
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
		logAuthInternalError("passkey issue token", err)
		writePublicAuthError(w, auth.ErrOperationFailed)
		return
	}
	_ = s.auth.WriteAuditLog(r.Context(), userID, "login", r.Header.Get("X-Forwarded-For"), r.Header.Get("User-Agent"), `{"method":"passkey"}`)
	setRefreshCookie(w, pair.RefreshToken, s.cfg.TLS)
	writeJSON(w, http.StatusOK, "ok", authResponse{User: user, Token: pair.AccessToken})
}

func (s *Server) handlePasskeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if s.passkey == nil {
		writeError(w, http.StatusServiceUnavailable, auth.ErrOperationFailed.Error())
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
		writeError(w, http.StatusServiceUnavailable, auth.ErrOperationFailed.Error())
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
