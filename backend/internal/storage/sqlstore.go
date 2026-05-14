package storage

import (
	"context"
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
	Imports    map[string]storedImport      `json:"imports"`
	Identities map[string]msglayer.Identity `json:"identities"`
	Events     map[string]storedEvent       `json:"events"`
}

type storedImport struct {
	ID            string `json:"id"`
	SchemaVersion string `json:"schema_version"`
	ImportedAt    string `json:"imported_at"`
	SourcePath    string `json:"source_path"`
	RawJSON       string `json:"raw_json"`
}

type storedEvent struct {
	Item msglayer.TimelineItem `json:"item"`
	Raw  msglayer.Event        `json:"raw"`
}

type fileStore struct {
	mu       sync.RWMutex
	name     string
	path     string
	snapshot storeSnapshot
}

func newSQLStore(path, name string) Provider {
	return &fileStore{
		name: name,
		path: path,
		snapshot: storeSnapshot{
			Imports:    map[string]storedImport{},
			Identities: map[string]msglayer.Identity{},
			Events:     map[string]storedEvent{},
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

func (s *fileStore) Import(ctx context.Context, sourcePath string, export msglayer.RootExport, raw []byte) (string, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	importID := fmt.Sprintf("import_%d", time.Now().UnixNano())
	s.snapshot.Imports[importID] = storedImport{
		ID:            importID,
		SchemaVersion: export.Version,
		ImportedAt:    time.Now().UTC().Format(time.RFC3339),
		SourcePath:    sourcePath,
		RawJSON:       string(raw),
	}
	for _, identity := range export.Identities {
		s.snapshot.Identities[identity.ID] = identity
	}
	for _, event := range export.Events {
		s.snapshot.Events[event.ID] = storedEvent{
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

func (s *fileStore) ExportImport(ctx context.Context, importID string) ([]byte, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	if importID == "" {
		var err error
		importID, err = s.latestImportIDLocked()
		if err != nil {
			return nil, err
		}
	}
	item, ok := s.snapshot.Imports[importID]
	if !ok {
		return nil, fmt.Errorf("import not found: %s", importID)
	}
	return []byte(item.RawJSON), nil
}

func (s *fileStore) LatestImportID(ctx context.Context) (string, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latestImportIDLocked()
}

func (s *fileStore) latestImportIDLocked() (string, error) {
	if len(s.snapshot.Imports) == 0 {
		return "", fmt.Errorf("no imports found")
	}
	imports := make([]storedImport, 0, len(s.snapshot.Imports))
	for _, item := range s.snapshot.Imports {
		imports = append(imports, item)
	}
	sort.Slice(imports, func(i, j int) bool { return imports[i].ImportedAt > imports[j].ImportedAt })
	return imports[0].ID, nil
}

func (s *fileStore) GetEvent(ctx context.Context, id string) (msglayer.TimelineItem, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.snapshot.Events[id]
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

func (s *fileStore) ListIdentities(ctx context.Context) ([]msglayer.Identity, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]msglayer.Identity, 0, len(s.snapshot.Identities))
	for _, identity := range s.snapshot.Identities {
		items = append(items, identity)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].DisplayName < items[j].DisplayName })
	return items, nil
}

func (s *fileStore) GetIdentity(ctx context.Context, id string) (msglayer.Identity, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	identity, ok := s.snapshot.Identities[id]
	if !ok {
		return msglayer.Identity{}, fmt.Errorf("identity not found: %s", id)
	}
	return identity, nil
}

func (s *fileStore) GetThread(ctx context.Context, threadID string) ([]msglayer.TimelineItem, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	var items []msglayer.TimelineItem
	for _, stored := range s.snapshot.Events {
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
