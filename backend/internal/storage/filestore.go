package storage

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
)

type storeSnapshot struct {
	Users         map[string]UserRecord         `json:"users"`
	RefreshTokens map[string]RefreshTokenRecord `json:"refresh_tokens"`
	Sessions      map[string]SessionRecord      `json:"sessions"`
	AuditLogs     map[string]AuditRecord        `json:"audit_logs"`
	Passkeys      map[string]PasskeyCredential  `json:"passkeys"`
	Challenges    map[string]ChallengeRecord    `json:"challenges"`
	AuthMethods   map[string]AuthMethodRecord   `json:"auth_methods"`
	Imports       map[string]storedImport       `json:"imports"`
	Identities    map[string]storedIdentity     `json:"identities"`
	Events        map[string]storedEvent        `json:"events"`
	Setup         *SetupRecord                  `json:"setup,omitempty"`
}

type storedImport struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	SchemaVersion string `json:"schema_version"`
	ImportedAt    string `json:"imported_at"`
	SourcePath    string `json:"source_path"`
	RawJSON       string `json:"raw_json"`
	EventCount    int    `json:"event_count"`
	IdentityCount int    `json:"identity_count"`
}

type storedIdentity struct {
	UserID   string            `json:"user_id"`
	Identity msglayer.Identity `json:"identity"`
}

type storedEvent struct {
	UserID   string                `json:"user_id"`
	ImportID string                `json:"import_id"`
	Item     msglayer.TimelineItem `json:"item"`
	Raw      msglayer.Event        `json:"raw"`
}

type fileStore struct {
	mu       sync.RWMutex
	name     string
	path     string
	snapshot storeSnapshot
}

func newFileStore(path, name string) Provider {
	return &fileStore{
		name: name,
		path: path,
		snapshot: storeSnapshot{
			Users:         map[string]UserRecord{},
			RefreshTokens: map[string]RefreshTokenRecord{},
			Sessions:      map[string]SessionRecord{},
			AuditLogs:     map[string]AuditRecord{},
			Passkeys:      map[string]PasskeyCredential{},
			Challenges:    map[string]ChallengeRecord{},
			AuthMethods:   map[string]AuthMethodRecord{},
			Imports:       map[string]storedImport{},
			Identities:    map[string]storedIdentity{},
			Events:        map[string]storedEvent{},
		},
	}
}

func NewFileStoreProvider(path string) (Provider, error) {
	return newFileStore(path, "filestore"), nil
}

func (s *fileStore) Name() string { return s.name }

func (s *fileStore) Close() error { return nil }

func (s *fileStore) Init(ctx context.Context) error {
	_ = ctx
	if err := EnsureParentDir(s.path); err != nil {
		return err
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return s.persist()
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, &s.snapshot); err != nil {
		return err
	}
	s.initMaps()
	return nil
}

func (s *fileStore) initMaps() {
	if s.snapshot.Users == nil {
		s.snapshot.Users = map[string]UserRecord{}
	}
	if s.snapshot.RefreshTokens == nil {
		s.snapshot.RefreshTokens = map[string]RefreshTokenRecord{}
	}
	if s.snapshot.Sessions == nil {
		s.snapshot.Sessions = map[string]SessionRecord{}
	}
	if s.snapshot.AuditLogs == nil {
		s.snapshot.AuditLogs = map[string]AuditRecord{}
	}
	if s.snapshot.Passkeys == nil {
		s.snapshot.Passkeys = map[string]PasskeyCredential{}
	}
	if s.snapshot.Challenges == nil {
		s.snapshot.Challenges = map[string]ChallengeRecord{}
	}
	if s.snapshot.AuthMethods == nil {
		s.snapshot.AuthMethods = map[string]AuthMethodRecord{}
	}
	if s.snapshot.Imports == nil {
		s.snapshot.Imports = map[string]storedImport{}
	}
	if s.snapshot.Identities == nil {
		s.snapshot.Identities = map[string]storedIdentity{}
	}
	if s.snapshot.Events == nil {
		s.snapshot.Events = map[string]storedEvent{}
	}
}
