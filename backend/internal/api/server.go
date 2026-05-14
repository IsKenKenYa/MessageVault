package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/IsKenKenYa/Commory/backend/internal/importers"
	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
	"github.com/IsKenKenYa/Commory/backend/internal/query"
	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

type Server struct {
	store     storage.Provider
	service   query.Service
	validator *msglayer.Validator
	importer  importers.Importer
}

func NewServer(store storage.Provider, validator *msglayer.Validator) *Server {
	return &Server{
		store:     store,
		service:   query.New(store),
		validator: validator,
		importer:  importers.JSONImporter{},
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/imports", s.handleImports)
	mux.HandleFunc("/api/validate", s.handleValidate)
	mux.HandleFunc("/api/events", s.handleEvents)
	mux.HandleFunc("/api/events/", s.handleEvent)
	mux.HandleFunc("/api/timeline", s.handleTimeline)
	mux.HandleFunc("/api/identities", s.handleIdentities)
	mux.HandleFunc("/api/identities/", s.handleIdentity)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/threads/", s.handleThread)
	return mux
}

func (s *Server) handleImports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	export, raw, err := s.importer.Import(req.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.validator.ValidateBytes(raw); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	importID, err := s.store.Import(r.Context(), req.Path, export, raw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"import_id":        importID,
		"msglayer_version": export.Version,
	})
}

func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.validator.ValidateFile(req.Path); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"valid": true})
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	items, err := s.service.Events(r.Context(), buildSearchParams(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleEvent(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/events/"):]
	item, err := s.service.Event(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleTimeline(w http.ResponseWriter, r *http.Request) {
	items, err := s.service.Timeline(r.Context(), buildSearchParams(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleIdentities(w http.ResponseWriter, r *http.Request) {
	items, err := s.service.Identities(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleIdentity(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/identities/"):]
	item, err := s.service.Identity(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	items, err := s.service.Search(r.Context(), buildSearchParams(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleThread(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/threads/"):]
	items, err := s.service.Thread(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func buildSearchParams(r *http.Request) msglayer.SearchParams {
	q := r.URL.Query()
	return msglayer.SearchParams{
		Keyword:     q.Get("q"),
		ContactID:   q.Get("contact"),
		Type:        q.Get("type"),
		Participant: q.Get("participant"),
		From:        q.Get("from"),
		To:          q.Get("to"),
		Limit:       100,
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func Shutdown(ctx context.Context, srv *http.Server) error {
	return srv.Shutdown(ctx)
}
