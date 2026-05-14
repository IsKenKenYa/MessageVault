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
	Imports       map[string]storedImport       `json:"imports"`
	Identities    map[string]storedIdentity     `json:"identities"`
	Events        map[string]storedEvent        `json:"events"`
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
	UserID   string               `json:"user_id"`
	ImportID string               `json:"import_id"`
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
			Imports:       map[string]storedImport{},
			Identities:    map[string]storedIdentity{},
			Events:        map[string]storedEvent{},
		},
	}
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
	return json.Unmarshal(data, &s.snapshot)
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
	item, ok := s.snapshot.Events[scopeKey(userID, id)]
	if !ok || item.UserID != userID {
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
	identity, ok := s.snapshot.Identities[scopeKey(userID, id)]
	if !ok || identity.UserID != userID {
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
	return fmt.Sprintf("%x:%s", hash[:4], rawID)
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
