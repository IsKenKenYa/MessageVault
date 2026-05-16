package api

import (
	"encoding/json"
	"net/http"

	"github.com/IsKenKenYa/Commory/backend/internal/setup"
)

func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetSetup(w, r)
	case http.MethodPost:
		s.handlePostSetup(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleGetSetup(w http.ResponseWriter, r *http.Request) {
	status, err := s.setupSvc.GetStatus(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "ok", status)
}

func (s *Server) handlePostSetup(w http.ResponseWriter, r *http.Request) {
	var req setup.SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.setupSvc.Initialize(r.Context(), req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "initialized", map[string]any{"success": true})
}
