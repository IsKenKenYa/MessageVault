package api

import (
	"net/http"

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
