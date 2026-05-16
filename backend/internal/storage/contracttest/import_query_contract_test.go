package contracttest

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
	"github.com/google/uuid"
)

func testImportAndQuery(t *testing.T, factory ProviderFactory) {
	store := factory(t)
	defer store.Close()
	ctx := context.Background()

	userID := "test-user-" + uuid.New().String()[:8]
	export := msglayer.RootExport{
		Version:    "msglayer/v0.1",
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Source:     msglayer.Source{Platform: "android", DeviceID: "test"},
		Identities: []msglayer.Identity{
			{ID: "id_1", Type: "person", DisplayName: "张三", Phones: []string{"+8613800000001"}},
			{ID: "id_2", Type: "person", DisplayName: "李四", Phones: []string{"+8613800000002"}},
		},
		Events: []msglayer.Event{
			{
				ID:           "evt_1",
				Type:         "sms",
				Timestamp:    time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339),
				Direction:    "outbound",
				Participants: []string{"id_1"},
				Content:      map[string]any{"text": "你好，验证码是123456"},
			},
			{
				ID:           "evt_2",
				Type:         "call",
				Timestamp:    time.Now().UTC().Format(time.RFC3339),
				Direction:    "inbound",
				Participants: []string{"id_2"},
				Content:      map[string]any{"duration_sec": 120, "call_type": "incoming"},
			},
		},
	}
	raw, _ := json.Marshal(export)

	importID, err := store.Import(ctx, userID, "test.json", export, raw)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if importID == "" {
		t.Fatal("expected import ID")
	}

	imports, err := store.ListImports(ctx, userID)
	if err != nil {
		t.Fatalf("ListImports: %v", err)
	}
	if len(imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(imports))
	}
	if imports[0].EventCount != 2 {
		t.Fatalf("expected 2 events, got %d", imports[0].EventCount)
	}

	exported, err := store.ExportImport(ctx, userID, importID)
	if err != nil {
		t.Fatalf("ExportImport: %v", err)
	}
	if len(exported) == 0 {
		t.Fatal("expected exported data")
	}

	event, err := store.GetEvent(ctx, userID, "evt_1")
	if err != nil {
		t.Fatalf("GetEvent: %v", err)
	}
	if event.Type != "sms" {
		t.Fatalf("expected sms type, got %s", event.Type)
	}

	results, err := store.Search(ctx, msglayer.SearchParams{UserID: userID, Keyword: "验证码", Limit: 10})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results for '验证码'")
	}

	timeline, err := store.Timeline(ctx, msglayer.SearchParams{UserID: userID, Limit: 10})
	if err != nil {
		t.Fatalf("Timeline: %v", err)
	}
	if len(timeline) != 2 {
		t.Fatalf("expected 2 timeline items, got %d", len(timeline))
	}

	identities, err := store.ListIdentities(ctx, userID)
	if err != nil {
		t.Fatalf("ListIdentities: %v", err)
	}
	if len(identities) != 2 {
		t.Fatalf("expected 2 identities, got %d", len(identities))
	}

	identity, err := store.GetIdentity(ctx, userID, "id_1")
	if err != nil {
		t.Fatalf("GetIdentity: %v", err)
	}
	if identity.DisplayName != "张三" {
		t.Fatalf("expected 张三, got %s", identity.DisplayName)
	}

	thread, err := store.GetThread(ctx, userID, "thread_42")
	if err != nil {
		t.Fatalf("GetThread: %v", err)
	}
	_ = thread
}

func testIdentityCRUD(t *testing.T, factory ProviderFactory) {
	store := factory(t)
	defer store.Close()
	ctx := context.Background()

	userID := "crud-user-" + uuid.New().String()[:8]
	identity := msglayer.Identity{
		ID:          "identity_" + uuid.New().String()[:8],
		Type:        "person",
		DisplayName: "测试用户",
		Phones:      []string{"+8613800000000"},
		Emails:      []string{"test@example.com"},
		Labels:      []string{"朋友"},
		Meta:        map[string]any{"source": "test"},
	}

	export := msglayer.RootExport{
		Version:    "msglayer/v0.1",
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Identities: []msglayer.Identity{identity},
		Events:     []msglayer.Event{},
	}
	raw, _ := json.Marshal(export)

	if _, err := store.Import(ctx, userID, "", export, raw); err != nil {
		t.Fatalf("Import: %v", err)
	}

	got, err := store.GetIdentity(ctx, userID, identity.ID)
	if err != nil {
		t.Fatalf("GetIdentity: %v", err)
	}
	if got.DisplayName != "测试用户" {
		t.Fatalf("display name mismatch: got %s", got.DisplayName)
	}
	if len(got.Phones) != 1 || got.Phones[0] != "+8613800000000" {
		t.Fatalf("phones mismatch: %v", got.Phones)
	}
}

func testSearchPagination(t *testing.T, factory ProviderFactory) {
	store := factory(t)
	defer store.Close()
	ctx := context.Background()

	userID := "page-user-" + uuid.New().String()[:8]
	var events []msglayer.Event
	for i := 0; i < 25; i++ {
		events = append(events, msglayer.Event{
			ID:        fmt.Sprintf("page_evt_%d", i),
			Type:      "sms",
			Timestamp: time.Now().Add(-time.Duration(25-i) * time.Minute).UTC().Format(time.RFC3339),
			Direction: "outbound",
			Content:   map[string]any{"text": fmt.Sprintf("消息编号 %d", i)},
		})
	}

	export := msglayer.RootExport{
		Version:    "msglayer/v0.1",
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Events:     events,
	}
	raw, _ := json.Marshal(export)

	if _, err := store.Import(ctx, userID, "", export, raw); err != nil {
		t.Fatalf("Import: %v", err)
	}

	page1, err := store.ListEvents(ctx, msglayer.SearchParams{UserID: userID, Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListEvents page1: %v", err)
	}
	if len(page1) != 10 {
		t.Fatalf("expected 10 items, got %d", len(page1))
	}

	page2, err := store.ListEvents(ctx, msglayer.SearchParams{UserID: userID, Limit: 10, Offset: 10})
	if err != nil {
		t.Fatalf("ListEvents page2: %v", err)
	}
	if len(page2) != 10 {
		t.Fatalf("expected 10 items, got %d", len(page2))
	}

	page3, err := store.ListEvents(ctx, msglayer.SearchParams{UserID: userID, Limit: 10, Offset: 20})
	if err != nil {
		t.Fatalf("ListEvents page3: %v", err)
	}
	if len(page3) != 5 {
		t.Fatalf("expected 5 items, got %d", len(page3))
	}

	if page1[0].EventID == page2[0].EventID {
		t.Fatal("page1 and page2 should not overlap")
	}
}
