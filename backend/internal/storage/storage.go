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
	ParentID  string    `json:"parent_id"`
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

type SetupStatus struct {
	Initialized  bool   `json:"initialized"`
	Version      string `json:"version"`
	DatabaseType string `json:"database_type"`
}

type SetupRecord struct {
	ID            string `json:"id"`
	Version       string `json:"version"`
	InitializedAt string `json:"initialized_at"`
	UsageMode     string `json:"usage_mode"`
}

type SessionRecord struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	RefreshTokenID string    `json:"refresh_token_id"`
	DeviceName     string    `json:"device_name"`
	DeviceType     string    `json:"device_type"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	CreatedAt      time.Time `json:"created_at"`
	LastSeenAt     time.Time `json:"last_seen_at"`
}

type AuditRecord struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

type PasskeyCredential struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	CredentialID    string    `json:"credential_id"`
	PublicKey       string    `json:"public_key"`
	AttestationType string    `json:"attestation_type"`
	AAGUID          string    `json:"aaguid"`
	SignCount       uint32    `json:"sign_count"`
	Transports      string    `json:"transports"`
	Name            string    `json:"name"`
	LastUsedAt      time.Time `json:"last_used_at"`
	CreatedAt       time.Time `json:"created_at"`
}

type ChallengeRecord struct {
	ID        string    `json:"id"`
	Challenge string    `json:"challenge"`
	UserID    string    `json:"user_id"`
	FlowType  string    `json:"flow_type"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthMethodRecord struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	ProviderType   string    `json:"provider_type"`
	ProviderUserID string    `json:"provider_user_id"`
	Metadata       string    `json:"metadata"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
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
	GetSetupStatus(context.Context) (SetupStatus, error)
	SaveSetup(context.Context, SetupRecord) error
	HasAdminUser(context.Context) (bool, error)
	UpdateUserPasswordHash(context.Context, string, string, string) error // userID, newHash, newSalt

	// 重放检测：查找任何匹配的 token（包括已撤销的）
	FindAnyRefreshTokenByHash(context.Context, string) (RefreshTokenRecord, error)
	// 令牌族撤销：按 parent_id 链撤销整族
	RevokeRefreshTokenFamily(context.Context, string) error
	RevokeRefreshTokenByID(context.Context, string) error

	// 会话管理
	CreateSession(context.Context, SessionRecord) error
	GetSession(context.Context, string) (SessionRecord, error)
	GetSessionByRefreshTokenID(context.Context, string) (SessionRecord, error)
	ListSessionsByUser(context.Context, string) ([]SessionRecord, error)
	RevokeSession(context.Context, string) error
	RevokeOtherSessions(context.Context, string, string) error // userID, currentSessionID
	UpdateSessionLastSeen(context.Context, string) error
	UpdateSessionRefreshToken(context.Context, string, string) error // sessionID, refreshTokenID

	// 审计日志
	CreateAuditLog(context.Context, AuditRecord) error
	ListAuditLogs(context.Context, string, string, int, int) ([]AuditRecord, error) // userID, action, limit, offset
	CountAuditLogs(context.Context, string, string) (int, error)

	// Passkey
	CreatePasskeyCredential(context.Context, PasskeyCredential) error
	GetPasskeyByCredentialID(context.Context, string) (PasskeyCredential, error)
	ListPasskeysByUser(context.Context, string) ([]PasskeyCredential, error)
	UpdatePasskeyLastUsed(context.Context, string, uint32) error // id, signCount
	DeletePasskey(context.Context, string, string) error         // id, userID

	// 认证 Challenge
	CreateChallenge(context.Context, ChallengeRecord) error
	GetChallenge(context.Context, string) (ChallengeRecord, error)
	DeleteChallenge(context.Context, string) error

	// 统一认证方式
	CreateAuthMethod(context.Context, AuthMethodRecord) error
	GetAuthMethodByProvider(context.Context, string, string) (AuthMethodRecord, error) // providerType, providerUserID

	Close() error
}
