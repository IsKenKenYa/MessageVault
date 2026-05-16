package api

import (
	"encoding/json"
	"net/http"

	"github.com/IsKenKenYa/Commory/backend/internal/auth"
)

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		UserName string `json:"userName"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writePublicAuthError(w, auth.ErrInvalidRequest)
		return
	}
	user, pair, err := s.auth.RegisterWithDevice(r.Context(), req.UserName, req.Email, req.Password,
		r.Header.Get("X-Commory-Device"), r.Header.Get("X-Forwarded-For"), r.Header.Get("User-Agent"))
	if err != nil {
		if auth.PublicErrorCode(err) == auth.ErrOperationFailed.Error() {
			logAuthInternalError("register", err)
		}
		writePublicAuthError(w, err)
		return
	}
	setRefreshCookie(w, pair.RefreshToken, s.cfg.TLS)
	writeJSON(w, http.StatusCreated, "registered", authResponse{User: user, Token: pair.AccessToken})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		UserName string `json:"userName"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writePublicAuthError(w, auth.ErrInvalidRequest)
		return
	}
	user, pair, err := s.auth.LoginWithDevice(r.Context(), r.RemoteAddr, req.UserName, req.Password,
		r.Header.Get("X-Commory-Device"), r.Header.Get("X-Forwarded-For"), r.Header.Get("User-Agent"))
	if err != nil {
		if auth.PublicErrorCode(err) == auth.ErrOperationFailed.Error() {
			logAuthInternalError("login", err)
		}
		writePublicAuthError(w, err)
		return
	}
	setRefreshCookie(w, pair.RefreshToken, s.cfg.TLS)
	writeJSON(w, http.StatusOK, "ok", authResponse{User: user, Token: pair.AccessToken})
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	refreshToken := ""
	if cookie, err := r.Cookie(refreshCookieName); err == nil {
		refreshToken = cookie.Value
	}
	pair, err := s.auth.Refresh(r.Context(), refreshToken)
	if err != nil {
		if err != auth.ErrRefreshTokenRetry {
			clearRefreshCookie(w, s.cfg.TLS)
		}
		if auth.PublicErrorCode(err) == auth.ErrOperationFailed.Error() {
			logAuthInternalError("refresh", err)
		}
		writePublicAuthError(w, err)
		return
	}
	setRefreshCookie(w, pair.RefreshToken, s.cfg.TLS)
	writeJSON(w, http.StatusOK, "ok", tokenResponse{Token: pair.AccessToken})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	refreshToken := req.RefreshToken
	if refreshToken == "" {
		if cookie, err := r.Cookie(refreshCookieName); err == nil {
			refreshToken = cookie.Value
		}
	}
	if refreshToken != "" {
		if err := s.auth.RevokeRefreshToken(r.Context(), refreshToken); err != nil {
			logAuthInternalError("logout revoke", err)
		}
	}
	clearRefreshCookie(w, s.cfg.TLS)
	writeJSON(w, http.StatusOK, "logged out", map[string]any{"success": true})
}

func (s *Server) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	user, err := s.auth.UserInfo(r.Context(), userID)
	if err != nil {
		if auth.PublicErrorCode(err) == auth.ErrOperationFailed.Error() {
			logAuthInternalError("user info", err)
		}
		writePublicAuthError(w, auth.ErrUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, "ok", user)
}
