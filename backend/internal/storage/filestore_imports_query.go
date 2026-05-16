package storage

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
)

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
