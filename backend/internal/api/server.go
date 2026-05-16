package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	urlpath "path"
	"path/filepath"
	"strings"

	"github.com/IsKenKenYa/Commory/backend/internal/auth"
	"github.com/IsKenKenYa/Commory/backend/internal/auth/oauth"
	"github.com/IsKenKenYa/Commory/backend/internal/auth/passkey"
	"github.com/IsKenKenYa/Commory/backend/internal/config"
	"github.com/IsKenKenYa/Commory/backend/internal/importers"
	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
	"github.com/IsKenKenYa/Commory/backend/internal/query"
	"github.com/IsKenKenYa/Commory/backend/internal/setup"
	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

const refreshCookieName = "commory_refresh_token"

type Server struct {
	cfg       config.Config
	store     storage.Provider
	service   query.Service
	auth      *auth.Service
	passkey   *passkey.Service
	oauth     *oauth.Registry
	validator *msglayer.Validator
	importer  importers.Importer
	setupSvc  *setup.Service
}

func NewServer(cfg config.Config, store storage.Provider, validator *msglayer.Validator) *Server {
	authSvc := auth.NewService(store, cfg.AuthSecret)

	var passkeySvc *passkey.Service
	if cfg.PasskeyRPID != "" && cfg.PasskeyOrigin != "" {
		passkeySvc, _ = passkey.NewService(store, cfg.PasskeyRPName, cfg.PasskeyRPID, cfg.PasskeyOrigin)
	}

	return &Server{
		cfg:       cfg,
		store:     store,
		service:   query.New(store),
		auth:      authSvc,
		passkey:   passkeySvc,
		oauth:     oauth.NewRegistry(),
		validator: validator,
		importer:  importers.JSONImporter{},
		setupSvc:  setup.NewService(store, authSvc),
	}
}

func (s *Server) Handler() http.Handler {
	publicMux := http.NewServeMux()
	publicMux.HandleFunc("/api/auth/register", s.handleRegister)
	publicMux.HandleFunc("/api/auth/login", s.handleLogin)
	publicMux.HandleFunc("/api/auth/refresh", s.handleRefresh)
	publicMux.HandleFunc("/api/auth/logout", s.handleLogout)
	publicMux.HandleFunc("/api/setup", s.handleSetup)
	publicMux.HandleFunc("/api/auth/passkey/login/begin", s.handlePasskeyLoginBegin)
	publicMux.HandleFunc("/api/auth/passkey/login/finish", s.handlePasskeyLoginFinish)
	publicMux.HandleFunc("/api/oauth/state", s.handleOAuthState)
	publicMux.HandleFunc("/api/oauth/providers", s.handleOAuthProviders)
	publicMux.HandleFunc("/api/oauth/", s.handleOAuthCallback)

	privateMux := http.NewServeMux()
	privateMux.HandleFunc("/api/user/info", s.handleUserInfo)
	privateMux.HandleFunc("/api/imports", s.handleImports)
	privateMux.HandleFunc("/api/imports/", s.handleImportExport)
	privateMux.HandleFunc("/api/imports/upload", s.handleImportUpload)
	privateMux.HandleFunc("/api/imports/path", s.handleImportPath)
	privateMux.HandleFunc("/api/validate/upload", s.handleValidateUpload)
	privateMux.HandleFunc("/api/validate/path", s.handleValidatePath)
	privateMux.HandleFunc("/api/dashboard", s.handleDashboard)
	privateMux.HandleFunc("/api/events", s.handleEvents)
	privateMux.HandleFunc("/api/events/", s.handleEvent)
	privateMux.HandleFunc("/api/timeline", s.handleTimeline)
	privateMux.HandleFunc("/api/identities", s.handleIdentities)
	privateMux.HandleFunc("/api/identities/", s.handleIdentity)
	privateMux.HandleFunc("/api/search", s.handleSearch)
	privateMux.HandleFunc("/api/threads/", s.handleThread)
	privateMux.HandleFunc("/api/sessions", s.handleSessions)
	privateMux.HandleFunc("/api/sessions/", s.handleSession)
	privateMux.HandleFunc("/api/audit-log", s.handleAuditLogs)

	root := http.NewServeMux()
	root.Handle("/api/auth/passkey/register/begin", auth.Middleware(s.auth, http.HandlerFunc(s.handlePasskeyRegisterBegin)))
	root.Handle("/api/auth/passkey/register/finish", auth.Middleware(s.auth, http.HandlerFunc(s.handlePasskeyRegisterFinish)))
	root.Handle("/api/auth/passkey", auth.Middleware(s.auth, http.HandlerFunc(s.handlePasskeys)))
	root.Handle("/api/auth/passkey/", auth.Middleware(s.auth, http.HandlerFunc(s.handlePasskeyDelete)))
	root.Handle("/api/auth/", publicMux)
	root.Handle("/api/setup", publicMux)
	root.Handle("/api/", auth.Middleware(s.auth, privateMux))
	if s.cfg.WebRoot != "" {
		root.Handle("/", spaHandler(s.cfg.WebRoot))
	}
	return root
}

func spaHandler(root string) http.Handler {
	fileServer := http.FileServer(http.Dir(root))
	indexPath := filepath.Join(root, "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		cleanPath := strings.TrimPrefix(urlpath.Clean("/"+r.URL.Path), "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}
		targetPath := filepath.Join(root, filepath.FromSlash(cleanPath))
		if info, err := os.Stat(targetPath); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, indexPath)
	})
}

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
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	user, pair, err := s.auth.RegisterWithDevice(r.Context(), req.UserName, req.Email, req.Password,
		r.Header.Get("X-Commory-Device"), r.Header.Get("X-Forwarded-For"), r.Header.Get("User-Agent"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	setRefreshCookie(w, pair.RefreshToken, s.cfg.TLS)
	writeJSON(w, http.StatusCreated, "registered", map[string]any{
		"user":         user,
		"token":        pair.AccessToken,
		"refreshToken": pair.RefreshToken,
	})
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
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	user, pair, err := s.auth.LoginWithDevice(r.Context(), r.RemoteAddr, req.UserName, req.Password,
		r.Header.Get("X-Commory-Device"), r.Header.Get("X-Forwarded-For"), r.Header.Get("User-Agent"))
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	setRefreshCookie(w, pair.RefreshToken, s.cfg.TLS)
	writeJSON(w, http.StatusOK, "ok", map[string]any{
		"user":         user,
		"token":        pair.AccessToken,
		"refreshToken": pair.RefreshToken,
	})
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
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
	pair, err := s.auth.Refresh(r.Context(), refreshToken)
	if err != nil {
		clearRefreshCookie(w, s.cfg.TLS)
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	setRefreshCookie(w, pair.RefreshToken, s.cfg.TLS)
	writeJSON(w, http.StatusOK, "ok", pair)
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
	_ = s.auth.RevokeRefreshToken(r.Context(), refreshToken)
	clearRefreshCookie(w, s.cfg.TLS)
	writeJSON(w, http.StatusOK, "logged out", map[string]any{"success": true})
}

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

func (s *Server) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	user, err := s.auth.UserInfo(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, "ok", user)
}

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

func readUpload(r *http.Request) ([]byte, string, error) {
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return nil, "", err
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			return nil, "", err
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		return data, header.Filename, err
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, "", err
	}
	return data, "upload.json", nil
}

func writeJSON(w http.ResponseWriter, status int, msg string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code": status,
		"msg":  msg,
		"data": data,
	})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, msg, nil)
}

func setRefreshCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   secure,
	})
}

func clearRefreshCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
	})
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

func Shutdown(ctx context.Context, srv *http.Server) error {
	return srv.Shutdown(ctx)
}
