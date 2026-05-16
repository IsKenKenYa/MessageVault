package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/IsKenKenYa/Commory/backend/internal/auth"
)

func (s *Server) handleImports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	items, err := s.store.ListImports(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "ok", items)
}

func (s *Server) handleImportExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/imports/")
	if !strings.HasSuffix(path, "/export") {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	importID := strings.TrimSuffix(path, "/export")
	importID = strings.TrimSuffix(importID, "/")
	raw, err := s.store.ExportImport(r.Context(), auth.UserIDFromContext(r.Context()), importID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", importID+".json"))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

func (s *Server) handleImportUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	raw, sourceName, err := readUpload(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	export, raw, err := s.importer.ImportBytes(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.validator.ValidateBytes(raw); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	importID, err := s.store.Import(r.Context(), userID, sourceName, export, raw)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, "imported", map[string]any{
		"import_id":        importID,
		"msglayer_version": export.Version,
	})
}

func (s *Server) handleImportPath(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := s.requireAdmin(r.Context()); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	resolved, err := s.resolveAllowedPath(req.Path)
	if err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	export, raw, err := s.importer.Import(resolved)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.validator.ValidateBytes(raw); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	importID, err := s.store.Import(r.Context(), auth.UserIDFromContext(r.Context()), resolved, export, raw)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, "imported", map[string]any{
		"import_id":        importID,
		"msglayer_version": export.Version,
	})
}

func (s *Server) handleValidateUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	raw, _, err := readUpload(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.validator.ValidateBytes(raw); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "valid", map[string]any{"valid": true})
}

func (s *Server) handleValidatePath(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := s.requireAdmin(r.Context()); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	resolved, err := s.resolveAllowedPath(req.Path)
	if err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	if err := s.validator.ValidateFile(resolved); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "valid", map[string]any{"valid": true})
}

func (s *Server) resolveAllowedPath(input string) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", fmt.Errorf("path is required")
	}
	resolved, err := filepath.Abs(input)
	if err != nil {
		return "", err
	}
	for _, root := range s.cfg.AllowedImportDir {
		if strings.HasPrefix(resolved, root+string(os.PathSeparator)) || resolved == root {
			return resolved, nil
		}
	}
	return "", fmt.Errorf("path is outside configured import roots")
}
