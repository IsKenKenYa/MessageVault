package storage

import (
	"context"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
)

type UserRecord struct {
	ID           string    `json:"id"`
	UserName     string    `json:"user_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	PasswordSalt string    `json:"password_salt"`
	Roles        []string  `json:"roles"`
	Buttons      []string  `json:"buttons"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RefreshTokenRecord struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	RevokedAt time.Time `json:"revoked_at"`
}

type ImportSummary struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	SchemaVersion string    `json:"schema_version"`
	ImportedAt    time.Time `json:"imported_at"`
	SourcePath    string    `json:"source_path"`
	EventCount    int       `json:"event_count"`
	IdentityCount int       `json:"identity_count"`
}

type Provider interface {
	Name() string
	Init(context.Context) error
	Import(context.Context, string, string, msglayer.RootExport, []byte) (string, error)
	ListImports(context.Context, string) ([]ImportSummary, error)
	ExportImport(context.Context, string, string) ([]byte, error)
	LatestImportID(context.Context, string) (string, error)
	GetEvent(context.Context, string, string) (msglayer.TimelineItem, error)
	ListEvents(context.Context, msglayer.SearchParams) ([]msglayer.TimelineItem, error)
	Search(context.Context, msglayer.SearchParams) ([]msglayer.TimelineItem, error)
	Timeline(context.Context, msglayer.SearchParams) ([]msglayer.TimelineItem, error)
	ListIdentities(context.Context, string) ([]msglayer.Identity, error)
	GetIdentity(context.Context, string, string) (msglayer.Identity, error)
	GetThread(context.Context, string, string) ([]msglayer.TimelineItem, error)
	CreateUser(context.Context, UserRecord) (UserRecord, error)
	FindUserByUserName(context.Context, string) (UserRecord, error)
	GetUser(context.Context, string) (UserRecord, error)
	SaveRefreshToken(context.Context, RefreshTokenRecord) error
	ConsumeRefreshToken(context.Context, string) (RefreshTokenRecord, error)
	Close() error
}
