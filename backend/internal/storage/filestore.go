package storage

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

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

func (s *fileStore) Import(ctx context.Context, userID string, sourcePath string, export msglayer.RootExport, raw []byte) (string, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	importID := fmt.Sprintf("import_%d", time.Now().UnixNano())
	s.snapshot.Imports[importID] = storedImport{
		ID:            importID,
		UserID:        userID,
		SchemaVersion: export.Version,
		ImportedAt:    time.Now().UTC().Format(time.RFC3339),
		SourcePath:    sourcePath,
		RawJSON:       string(raw),
		EventCount:    len(export.Events),
		IdentityCount: len(export.Identities),
	}
	for _, identity := range export.Identities {
		s.snapshot.Identities[scopeKey(userID, identity.ID)] = storedIdentity{
			UserID:   userID,
			Identity: identity,
		}
	}
	for _, event := range export.Events {
		s.snapshot.Events[scopeKey(userID, event.ID)] = storedEvent{
			UserID:   userID,
			ImportID: importID,
			Item: msglayer.TimelineItem{
				EventID:        event.ID,
				Type:           event.Type,
				Timestamp:      event.Timestamp,
				Direction:      event.Direction,
				ContentSummary: summarizeEvent(event),
				Participants:   append([]string(nil), event.Participants...),
				Meta:           event.Meta,
				SchemaVersion:  export.Version,
			},
			Raw: event,
		}
	}
	return importID, s.persist()
}

func (s *fileStore) ListImports(ctx context.Context, userID string) ([]ImportSummary, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]ImportSummary, 0, len(s.snapshot.Imports))
	for _, item := range s.snapshot.Imports {
		if item.UserID != userID {
			continue
		}
		importedAt, _ := time.Parse(time.RFC3339, item.ImportedAt)
		items = append(items, ImportSummary{
			ID:            item.ID,
			UserID:        item.UserID,
			SchemaVersion: item.SchemaVersion,
			ImportedAt:    importedAt,
			SourcePath:    item.SourcePath,
			EventCount:    item.EventCount,
			IdentityCount: item.IdentityCount,
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ImportedAt.After(items[j].ImportedAt) })
	return items, nil
}

func (s *fileStore) ExportImport(ctx context.Context, userID, importID string) ([]byte, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	if importID == "" {
		var err error
		importID, err = s.latestImportIDLocked(userID)
		if err != nil {
			return nil, err
		}
	}
	item, ok := s.snapshot.Imports[importID]
	if !ok || item.UserID != userID {
		return nil, fmt.Errorf("import not found: %s", importID)
	}
	return []byte(item.RawJSON), nil
}

func (s *fileStore) LatestImportID(ctx context.Context, userID string) (string, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latestImportIDLocked(userID)
}

func (s *fileStore) latestImportIDLocked(userID string) (string, error) {
	imports := make([]storedImport, 0, len(s.snapshot.Imports))
	for _, item := range s.snapshot.Imports {
		if item.UserID != userID {
			continue
		}
		imports = append(imports, item)
	}
	if len(imports) == 0 {
		return "", fmt.Errorf("no imports found")
	}
	sort.Slice(imports, func(i, j int) bool { return imports[i].ImportedAt > imports[j].ImportedAt })
	return imports[0].ID, nil
}

func (s *fileStore) GetEvent(ctx context.Context, userID, id string) (msglayer.TimelineItem, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.lookupScopedEventKey(s.snapshot.Events, userID, id)
	if !ok {
		return msglayer.TimelineItem{}, fmt.Errorf("event not found: %s", id)
	}
	return item.Item, nil
}

func (s *fileStore) ListEvents(ctx context.Context, params msglayer.SearchParams) ([]msglayer.TimelineItem, error) {
	return s.filterEvents(ctx, params, false)
}

func (s *fileStore) Search(ctx context.Context, params msglayer.SearchParams) ([]msglayer.TimelineItem, error) {
	return s.filterEvents(ctx, params, true)
}

func (s *fileStore) Timeline(ctx context.Context, params msglayer.SearchParams) ([]msglayer.TimelineItem, error) {
	return s.filterEvents(ctx, params, false)
}

func (s *fileStore) filterEvents(ctx context.Context, params msglayer.SearchParams, keywordOnly bool) ([]msglayer.TimelineItem, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	var items []msglayer.TimelineItem
	for _, stored := range s.snapshot.Events {
		if stored.UserID != params.UserID {
			continue
		}
		if keywordOnly && params.Keyword == "" {
			continue
		}
		if params.Keyword != "" && !matchesKeyword(stored, params.Keyword) {
			continue
		}
		if params.ContactID != "" && !contains(stored.Item.Participants, params.ContactID) {
			continue
		}
		if params.Participant != "" && !contains(stored.Item.Participants, params.Participant) {
			continue
		}
		if params.Type != "" && stored.Item.Type != params.Type {
			continue
		}
		// Timestamp ordering/filtering relies on normalized UTC RFC3339 strings.
		if params.From != "" && stored.Item.Timestamp < params.From {
			continue
		}
		if params.To != "" && stored.Item.Timestamp > params.To {
			continue
		}
		items = append(items, stored.Item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Timestamp > items[j].Timestamp })
	if params.Offset > 0 && params.Offset < len(items) {
		items = items[params.Offset:]
	} else if params.Offset >= len(items) {
		items = nil
	}
	if params.Limit > 0 && len(items) > params.Limit {
		items = items[:params.Limit]
	}
	return items, nil
}

func (s *fileStore) ListIdentities(ctx context.Context, userID string) ([]msglayer.Identity, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]msglayer.Identity, 0, len(s.snapshot.Identities))
	for _, identity := range s.snapshot.Identities {
		if identity.UserID != userID {
			continue
		}
		items = append(items, identity.Identity)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].DisplayName < items[j].DisplayName })
	return items, nil
}

func (s *fileStore) GetIdentity(ctx context.Context, userID, id string) (msglayer.Identity, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	identity, ok := s.lookupScopedKey(s.snapshot.Identities, userID, id)
	if !ok {
		return msglayer.Identity{}, fmt.Errorf("identity not found: %s", id)
	}
	return identity.Identity, nil
}

func (s *fileStore) GetThread(ctx context.Context, userID, threadID string) ([]msglayer.TimelineItem, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	var items []msglayer.TimelineItem
	for _, stored := range s.snapshot.Events {
		if stored.UserID != userID {
			continue
		}
		for _, relation := range stored.Raw.Relations {
			if relation.Type == "same_thread" && relation.Target == threadID {
				items = append(items, stored.Item)
				break
			}
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Timestamp < items[j].Timestamp })
	return items, nil
}

func (s *fileStore) CreateUser(ctx context.Context, user UserRecord) (UserRecord, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.snapshot.Users {
		if strings.EqualFold(existing.UserName, user.UserName) {
			return UserRecord{}, fmt.Errorf("username already exists")
		}
	}
	s.snapshot.Users[user.ID] = user
	return user, s.persist()
}

func (s *fileStore) FindUserByUserName(ctx context.Context, userName string) (UserRecord, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, user := range s.snapshot.Users {
		if strings.EqualFold(user.UserName, userName) {
			return user, nil
		}
	}
	return UserRecord{}, fmt.Errorf("user not found")
}

func (s *fileStore) GetUser(ctx context.Context, userID string) (UserRecord, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.snapshot.Users[userID]
	if !ok {
		return UserRecord{}, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *fileStore) SaveRefreshToken(ctx context.Context, token RefreshTokenRecord) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot.RefreshTokens[token.ID] = token
	return s.persist()
}

func (s *fileStore) ConsumeRefreshToken(ctx context.Context, tokenHash string) (RefreshTokenRecord, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, token := range s.snapshot.RefreshTokens {
		if token.TokenHash != tokenHash {
			continue
		}
		if !token.RevokedAt.IsZero() {
			return RefreshTokenRecord{}, fmt.Errorf("refresh token revoked")
		}
		if time.Now().After(token.ExpiresAt) {
			return RefreshTokenRecord{}, fmt.Errorf("refresh token expired")
		}
		token.RevokedAt = time.Now().UTC()
		s.snapshot.RefreshTokens[key] = token
		if err := s.persist(); err != nil {
			return RefreshTokenRecord{}, err
		}
		return token, nil
	}
	return RefreshTokenRecord{}, fmt.Errorf("refresh token not found")
}

func (s *fileStore) GetSetupStatus(ctx context.Context) (SetupStatus, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.snapshot.Setup != nil {
		return SetupStatus{Initialized: true, Version: s.snapshot.Setup.Version, DatabaseType: "sqlite"}, nil
	}
	hasAdmin, _ := s.hasAdminUserLocked()
	return SetupStatus{Initialized: hasAdmin, DatabaseType: "sqlite"}, nil
}

func (s *fileStore) SaveSetup(ctx context.Context, setup SetupRecord) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot.Setup = &setup
	return s.persist()
}

func (s *fileStore) HasAdminUser(ctx context.Context) (bool, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hasAdminUserLocked()
}

func (s *fileStore) hasAdminUserLocked() (bool, error) {
	for _, user := range s.snapshot.Users {
		for _, role := range user.Roles {
			if role == "R_ADMIN" || role == "R_SUPER" {
				return true, nil
			}
		}
	}
	return false, nil
}

func (s *fileStore) UpdateUserPasswordHash(ctx context.Context, userID, newHash, newSalt string) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.snapshot.Users[userID]
	if !ok {
		return fmt.Errorf("user not found")
	}
	user.PasswordHash = newHash
	user.PasswordSalt = newSalt
	user.UpdatedAt = time.Now().UTC()
	s.snapshot.Users[userID] = user
	return s.persist()
}

func (s *fileStore) FindAnyRefreshTokenByHash(_ context.Context, tokenHash string) (RefreshTokenRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, token := range s.snapshot.RefreshTokens {
		if token.TokenHash == tokenHash {
			return token, nil
		}
	}
	return RefreshTokenRecord{}, fmt.Errorf("refresh token not found")
}

func (s *fileStore) RevokeRefreshTokenFamily(_ context.Context, tokenID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue := []string{tokenID}
	seen := map[string]struct{}{}
	now := time.Now().UTC()
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if _, ok := seen[current]; ok {
			continue
		}
		seen[current] = struct{}{}
		if token, ok := s.snapshot.RefreshTokens[current]; ok && token.RevokedAt.IsZero() {
			token.RevokedAt = now
			s.snapshot.RefreshTokens[current] = token
		}
		for id, token := range s.snapshot.RefreshTokens {
			if token.ParentID == current {
				queue = append(queue, id)
			}
		}
	}
	return s.persist()
}

func (s *fileStore) RevokeRefreshTokenByID(_ context.Context, tokenID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	token, ok := s.snapshot.RefreshTokens[tokenID]
	if !ok {
		return nil
	}
	if token.RevokedAt.IsZero() {
		token.RevokedAt = time.Now().UTC()
		s.snapshot.RefreshTokens[tokenID] = token
	}
	return s.persist()
}

func (s *fileStore) CreateSession(_ context.Context, rec SessionRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot.Sessions[rec.ID] = rec
	return s.persist()
}

func (s *fileStore) GetSession(_ context.Context, sessionID string) (SessionRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.snapshot.Sessions[sessionID]
	if !ok {
		return SessionRecord{}, fmt.Errorf("session not found")
	}
	return session, nil
}

func (s *fileStore) GetSessionByRefreshTokenID(_ context.Context, refreshTokenID string) (SessionRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, session := range s.snapshot.Sessions {
		if session.RefreshTokenID == refreshTokenID {
			return session, nil
		}
	}
	return SessionRecord{}, fmt.Errorf("session not found")
}

func (s *fileStore) ListSessionsByUser(_ context.Context, userID string) ([]SessionRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]SessionRecord, 0, len(s.snapshot.Sessions))
	for _, session := range s.snapshot.Sessions {
		if session.UserID != userID {
			continue
		}
		items = append(items, session)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].LastSeenAt.After(items[j].LastSeenAt) })
	return items, nil
}

func (s *fileStore) RevokeSession(_ context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.snapshot.Sessions[sessionID]
	if !ok {
		return nil
	}
	session.LastSeenAt = time.Now().UTC()
	delete(s.snapshot.Sessions, sessionID)
	if session.RefreshTokenID != "" {
		if token, ok := s.snapshot.RefreshTokens[session.RefreshTokenID]; ok && token.RevokedAt.IsZero() {
			token.RevokedAt = time.Now().UTC()
			s.snapshot.RefreshTokens[session.RefreshTokenID] = token
		}
	}
	return s.persist()
}

func (s *fileStore) RevokeOtherSessions(_ context.Context, userID, currentSessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	for id, session := range s.snapshot.Sessions {
		if session.UserID != userID || id == currentSessionID {
			continue
		}
		if session.RefreshTokenID != "" {
			if token, ok := s.snapshot.RefreshTokens[session.RefreshTokenID]; ok && token.RevokedAt.IsZero() {
				token.RevokedAt = now
				s.snapshot.RefreshTokens[session.RefreshTokenID] = token
			}
		}
		delete(s.snapshot.Sessions, id)
	}
	return s.persist()
}

func (s *fileStore) UpdateSessionLastSeen(_ context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.snapshot.Sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found")
	}
	session.LastSeenAt = time.Now().UTC()
	s.snapshot.Sessions[sessionID] = session
	return s.persist()
}

func (s *fileStore) UpdateSessionRefreshToken(_ context.Context, sessionID, refreshTokenID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.snapshot.Sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found")
	}
	session.RefreshTokenID = refreshTokenID
	session.LastSeenAt = time.Now().UTC()
	s.snapshot.Sessions[sessionID] = session
	return s.persist()
}

func (s *fileStore) CreateAuditLog(_ context.Context, rec AuditRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot.AuditLogs[rec.ID] = rec
	return s.persist()
}

func (s *fileStore) ListAuditLogs(_ context.Context, userID, action string, limit, offset int) ([]AuditRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]AuditRecord, 0, len(s.snapshot.AuditLogs))
	for _, rec := range s.snapshot.AuditLogs {
		if userID != "" && rec.UserID != userID {
			continue
		}
		if action != "" && rec.Action != action {
			continue
		}
		items = append(items, rec)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	if offset > 0 && offset < len(items) {
		items = items[offset:]
	} else if offset >= len(items) {
		return nil, nil
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (s *fileStore) CountAuditLogs(ctx context.Context, userID, action string) (int, error) {
	items, err := s.ListAuditLogs(ctx, userID, action, 0, 0)
	if err != nil {
		return 0, err
	}
	return len(items), nil
}

func (s *fileStore) CreatePasskeyCredential(_ context.Context, rec PasskeyCredential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot.Passkeys[rec.ID] = rec
	return s.persist()
}

func (s *fileStore) GetPasskeyByCredentialID(_ context.Context, credentialID string) (PasskeyCredential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, passkey := range s.snapshot.Passkeys {
		if passkey.CredentialID == credentialID {
			return passkey, nil
		}
	}
	return PasskeyCredential{}, fmt.Errorf("passkey not found")
}

func (s *fileStore) ListPasskeysByUser(_ context.Context, userID string) ([]PasskeyCredential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]PasskeyCredential, 0, len(s.snapshot.Passkeys))
	for _, passkey := range s.snapshot.Passkeys {
		if passkey.UserID != userID {
			continue
		}
		items = append(items, passkey)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items, nil
}

func (s *fileStore) UpdatePasskeyLastUsed(_ context.Context, id string, signCount uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	passkey, ok := s.snapshot.Passkeys[id]
	if !ok {
		return fmt.Errorf("passkey not found")
	}
	passkey.SignCount = signCount
	passkey.LastUsedAt = time.Now().UTC()
	s.snapshot.Passkeys[id] = passkey
	return s.persist()
}

func (s *fileStore) DeletePasskey(_ context.Context, id, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	passkey, ok := s.snapshot.Passkeys[id]
	if !ok || passkey.UserID != userID {
		return nil
	}
	delete(s.snapshot.Passkeys, id)
	return s.persist()
}

func (s *fileStore) CreateChallenge(_ context.Context, rec ChallengeRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = time.Now().UTC()
	}
	s.snapshot.Challenges[rec.ID] = rec
	return s.persist()
}

func (s *fileStore) GetChallenge(_ context.Context, id string) (ChallengeRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	challenge, ok := s.snapshot.Challenges[id]
	if !ok || time.Now().UTC().After(challenge.ExpiresAt) {
		return ChallengeRecord{}, fmt.Errorf("challenge not found")
	}
	return challenge, nil
}

func (s *fileStore) DeleteChallenge(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.snapshot.Challenges, id)
	return s.persist()
}

func (s *fileStore) CreateAuthMethod(_ context.Context, rec AuthMethodRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.snapshot.AuthMethods {
		if existing.ProviderType == rec.ProviderType && existing.ProviderUserID == rec.ProviderUserID {
			return fmt.Errorf("auth method already exists")
		}
	}
	s.snapshot.AuthMethods[rec.ID] = rec
	return s.persist()
}

func (s *fileStore) GetAuthMethodByProvider(_ context.Context, providerType, providerUserID string) (AuthMethodRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, rec := range s.snapshot.AuthMethods {
		if rec.ProviderType == providerType && rec.ProviderUserID == providerUserID {
			return rec, nil
		}
	}
	return AuthMethodRecord{}, fmt.Errorf("auth method not found")
}

func EnsureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func (s *fileStore) persist() error {
	data, err := json.MarshalIndent(s.snapshot, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func scopeKey(userID, rawID string) string {
	hash := sha1.Sum([]byte(userID))
	return fmt.Sprintf("%x:%s", hash[:8], rawID)
}

func scopeKeyLegacy(userID, rawID string) string {
	hash := sha1.Sum([]byte(userID))
	return fmt.Sprintf("%x:%s", hash[:4], rawID)
}

func (s *fileStore) lookupScopedKey(m map[string]storedIdentity, userID, rawID string) (storedIdentity, bool) {
	key := scopeKey(userID, rawID)
	if v, ok := m[key]; ok && v.UserID == userID {
		return v, true
	}
	key = scopeKeyLegacy(userID, rawID)
	if v, ok := m[key]; ok && v.UserID == userID {
		return v, true
	}
	return storedIdentity{}, false
}

func (s *fileStore) lookupScopedEventKey(m map[string]storedEvent, userID, rawID string) (storedEvent, bool) {
	key := scopeKey(userID, rawID)
	if v, ok := m[key]; ok && v.UserID == userID {
		return v, true
	}
	key = scopeKeyLegacy(userID, rawID)
	if v, ok := m[key]; ok && v.UserID == userID {
		return v, true
	}
	return storedEvent{}, false
}

func summarizeEvent(event msglayer.Event) string {
	switch event.Type {
	case "sms":
		return asString(event.Content["text"])
	case "call":
		return fmt.Sprintf("%s %vsec", asString(event.Content["call_type"]), event.Content["duration_sec"])
	case "voice":
		if summary := asString(event.Content["summary"]); summary != "" {
			return summary
		}
		return asString(event.Content["transcript"])
	case "contact_snapshot":
		return asString(event.Content["identity_id"])
	default:
		return event.Type
	}
}

func matchesKeyword(event storedEvent, keyword string) bool {
	needle := strings.ToLower(keyword)
	if strings.Contains(strings.ToLower(event.Item.ContentSummary), needle) {
		return true
	}
	if transcript := asString(event.Raw.Content["transcript"]); strings.Contains(strings.ToLower(transcript), needle) {
		return true
	}
	if summary := asString(event.Raw.Content["summary"]); strings.Contains(strings.ToLower(summary), needle) {
		return true
	}
	return false
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	if value, ok := v.(string); ok {
		return value
	}
	return fmt.Sprintf("%v", v)
}
