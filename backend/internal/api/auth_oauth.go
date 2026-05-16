package api

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

func (s *Server) handleOAuthState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	state := randomHex(16)
	challengeID := randomHex(8)
	_ = s.store.CreateChallenge(r.Context(), storage.ChallengeRecord{
		ID:        challengeID,
		Challenge: state,
		FlowType:  "oauth_state",
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
	})
	writeJSON(w, http.StatusOK, "ok", map[string]any{
		"state":       state,
		"challengeId": challengeID,
	})
}

func (s *Server) handleOAuthProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	providers := s.oauth.ListEnabled()
	type providerInfo struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
	}
	result := make([]providerInfo, 0, len(providers))
	for _, p := range providers {
		result = append(result, providerInfo{Name: p.Name(), DisplayName: p.DisplayName()})
	}
	writeJSON(w, http.StatusOK, "ok", result)
}

func (s *Server) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	providerName := strings.TrimPrefix(r.URL.Path, "/api/oauth/")
	providerName = strings.TrimSuffix(providerName, "/")
	if providerName == "" || providerName == "state" || providerName == "providers" {
		writeError(w, http.StatusBadRequest, "provider name required")
		return
	}
	provider := s.oauth.Get(providerName)
	if provider == nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("oauth provider %q not found", providerName))
		return
	}
	writeJSON(w, http.StatusNotImplemented, "oauth callback not yet implemented", nil)
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
