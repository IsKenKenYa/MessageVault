package setup

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/auth"
	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

type Service struct {
	store storage.Provider
	auth  *auth.Service
}

func NewService(store storage.Provider, auth *auth.Service) *Service {
	return &Service{store: store, auth: auth}
}

type SetupStatusResponse struct {
	Status       bool   `json:"status"`
	DatabaseType string `json:"database_type"`
	RootInit     bool   `json:"root_init"`
}

func (s *Service) GetStatus(ctx context.Context) (SetupStatusResponse, error) {
	status, err := s.store.GetSetupStatus(ctx)
	if err != nil {
		return SetupStatusResponse{}, err
	}
	hasAdmin, _ := s.store.HasAdminUser(ctx)
	return SetupStatusResponse{
		Status:       status.Initialized,
		DatabaseType: status.DatabaseType,
		RootInit:     hasAdmin,
	}, nil
}

type SetupRequest struct {
	UserName        string `json:"userName"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	UsageMode       string `json:"usageMode"`
}

func (s *Service) Initialize(ctx context.Context, req SetupRequest) error {
	status, err := s.store.GetSetupStatus(ctx)
	if err != nil {
		return err
	}
	if status.Initialized {
		return fmt.Errorf("system already initialized")
	}

	hasAdmin, _ := s.store.HasAdminUser(ctx)
	if !hasAdmin {
		if strings.TrimSpace(req.UserName) == "" {
			return fmt.Errorf("admin username is required")
		}
		if len(req.Password) < 8 {
			return fmt.Errorf("password must be at least 8 characters")
		}
		if req.Password != req.ConfirmPassword {
			return fmt.Errorf("passwords do not match")
		}
		user, _, err := s.auth.RegisterAdmin(ctx, req.UserName, "", req.Password)
		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		_ = user
	}

	return s.store.SaveSetup(ctx, storage.SetupRecord{
		ID:            fmt.Sprintf("setup_%d", time.Now().UnixNano()),
		Version:       "1.0",
		InitializedAt: time.Now().UTC().Format(time.RFC3339),
		UsageMode:     req.UsageMode,
	})
}

func CheckSetup(ctx context.Context, store storage.Provider) {
	status, err := store.GetSetupStatus(ctx)
	if err != nil {
		return
	}
	if !status.Initialized {
		hasAdmin, _ := store.HasAdminUser(ctx)
		if hasAdmin {
			_ = store.SaveSetup(ctx, storage.SetupRecord{
				ID:            fmt.Sprintf("setup_%d", time.Now().UnixNano()),
				Version:       "0.9-legacy",
				InitializedAt: time.Now().UTC().Format(time.RFC3339),
				UsageMode:     "personal",
			})
		}
	}
}
