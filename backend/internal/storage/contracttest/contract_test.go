package contracttest

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
	"github.com/IsKenKenYa/Commory/backend/internal/storage"
	"github.com/google/uuid"
)

// ProviderFactory 创建一个临时的 Provider 实例。
type ProviderFactory func(t *testing.T) storage.Provider

// RunContractTests 对给定的 Provider 运行完整的存储契约测试套件。
func RunContractTests(t *testing.T, factory ProviderFactory) {
	t.Run("UserLifecycle", func(t *testing.T) { testUserLifecycle(t, factory) })
	t.Run("RefreshTokenRotation", func(t *testing.T) { testRefreshTokenRotation(t, factory) })
	t.Run("ImportAndQuery", func(t *testing.T) { testImportAndQuery(t, factory) })
	t.Run("IdentityCRUD", func(t *testing.T) { testIdentityCRUD(t, factory) })
	t.Run("SearchPagination", func(t *testing.T) { testSearchPagination(t, factory) })
	t.Run("SetupLifecycle", func(t *testing.T) { testSetupLifecycle(t, factory) })
	t.Run("SessionLifecycle", func(t *testing.T) { testSessionLifecycle(t, factory) })
	t.Run("AuditLog", func(t *testing.T) { testAuditLog(t, factory) })
	t.Run("PasskeyCredential", func(t *testing.T) { testPasskeyCredential(t, factory) })
	t.Run("ChallengeLifecycle", func(t *testing.T) { testChallengeLifecycle(t, factory) })
	t.Run("AuthMethod", func(t *testing.T) { testAuthMethod(t, factory) })
}

func testUserLifecycle(t *testing.T, factory ProviderFactory) {
	store := factory(t)
	defer store.Close()
	ctx := context.Background()

	user := storage.UserRecord{
		ID:           uuid.New().String(),
		UserName:     "testuser_" + uuid.New().String()[:8],
		Email:        "test@example.com",
		PasswordHash: "hash123",
		PasswordSalt: "salt123",
		Roles:        []string{"R_USER"},
		Buttons:      []string{"view", "import"},
	}

	created, err := store.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created.UserName != user.UserName {
		t.Fatalf("username mismatch: got %s", created.UserName)
	}

	found, err := store.FindUserByUserName(ctx, user.UserName)
	if err != nil {
		t.Fatalf("FindUserByUserName: %v", err)
	}
	if found.ID != user.ID {
		t.Fatalf("user ID mismatch: got %s", found.ID)
	}

	got, err := store.GetUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if got.Email != user.Email {
		t.Fatalf("email mismatch: got %s", got.Email)
	}

	hasAdmin, err := store.HasAdminUser(ctx)
	if err != nil {
		t.Fatalf("HasAdminUser: %v", err)
	}
	if hasAdmin {
		t.Fatal("expected no admin user")
	}

	// 创建重复用户名应失败
	_, err = store.CreateUser(ctx, storage.UserRecord{
		ID:       uuid.New().String(),
		UserName: user.UserName,
	})
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
}

func testRefreshTokenRotation(t *testing.T, factory ProviderFactory) {
	store := factory(t)
	defer store.Close()
	ctx := context.Background()

	user := storage.UserRecord{
		ID:           uuid.New().String(),
		UserName:     "tokenuser_" + uuid.New().String()[:8],
		PasswordHash: "hash",
		PasswordSalt: "salt",
		Roles:        []string{"R_USER"},
		Buttons:      []string{"view"},
	}
	if _, err := store.CreateUser(ctx, user); err != nil {
		t.Fatal(err)
	}

	token := storage.RefreshTokenRecord{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		TokenHash: "hash_" + uuid.New().String(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := store.SaveRefreshToken(ctx, token); err != nil {
		t.Fatalf("SaveRefreshToken: %v", err)
	}

	consumed, err := store.ConsumeRefreshToken(ctx, token.TokenHash)
	if err != nil {
		t.Fatalf("ConsumeRefreshToken: %v", err)
	}
	if consumed.UserID != user.ID {
		t.Fatalf("user ID mismatch: got %s", consumed.UserID)
	}

	// 二次消费应失败
	_, err = store.ConsumeRefreshToken(ctx, token.TokenHash)
	if err == nil {
		t.Fatal("expected error for already consumed token")
	}
}

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

	// 查询导入列表
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

	// 导出导入数据
	exported, err := store.ExportImport(ctx, userID, importID)
	if err != nil {
		t.Fatalf("ExportImport: %v", err)
	}
	if len(exported) == 0 {
		t.Fatal("expected exported data")
	}

	// 获取事件
	event, err := store.GetEvent(ctx, userID, "evt_1")
	if err != nil {
		t.Fatalf("GetEvent: %v", err)
	}
	if event.Type != "sms" {
		t.Fatalf("expected sms type, got %s", event.Type)
	}

	// 搜索
	results, err := store.Search(ctx, msglayer.SearchParams{UserID: userID, Keyword: "验证码", Limit: 10})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results for '验证码'")
	}

	// 时间线
	timeline, err := store.Timeline(ctx, msglayer.SearchParams{UserID: userID, Limit: 10})
	if err != nil {
		t.Fatalf("Timeline: %v", err)
	}
	if len(timeline) != 2 {
		t.Fatalf("expected 2 timeline items, got %d", len(timeline))
	}

	// 获取联系人
	identities, err := store.ListIdentities(ctx, userID)
	if err != nil {
		t.Fatalf("ListIdentities: %v", err)
	}
	if len(identities) != 2 {
		t.Fatalf("expected 2 identities, got %d", len(identities))
	}

	// 获取单个联系人
	identity, err := store.GetIdentity(ctx, userID, "id_1")
	if err != nil {
		t.Fatalf("GetIdentity: %v", err)
	}
	if identity.DisplayName != "张三" {
		t.Fatalf("expected 张三, got %s", identity.DisplayName)
	}

	// 线程查询
	thread, err := store.GetThread(ctx, userID, "thread_42")
	if err != nil {
		t.Fatalf("GetThread: %v", err)
	}
	_ = thread // 线程可能为空，因为测试数据没有 same_thread 关系
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

	// 第一页
	page1, err := store.ListEvents(ctx, msglayer.SearchParams{UserID: userID, Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListEvents page1: %v", err)
	}
	if len(page1) != 10 {
		t.Fatalf("expected 10 items, got %d", len(page1))
	}

	// 第二页
	page2, err := store.ListEvents(ctx, msglayer.SearchParams{UserID: userID, Limit: 10, Offset: 10})
	if err != nil {
		t.Fatalf("ListEvents page2: %v", err)
	}
	if len(page2) != 10 {
		t.Fatalf("expected 10 items, got %d", len(page2))
	}

	// 第三页
	page3, err := store.ListEvents(ctx, msglayer.SearchParams{UserID: userID, Limit: 10, Offset: 20})
	if err != nil {
		t.Fatalf("ListEvents page3: %v", err)
	}
	if len(page3) != 5 {
		t.Fatalf("expected 5 items, got %d", len(page3))
	}

	// 确保无重叠
	if page1[0].EventID == page2[0].EventID {
		t.Fatal("page1 and page2 should not overlap")
	}
}

func testSetupLifecycle(t *testing.T, factory ProviderFactory) {
	store := factory(t)
	defer store.Close()
	ctx := context.Background()

	status, err := store.GetSetupStatus(ctx)
	if err != nil {
		t.Fatalf("GetSetupStatus: %v", err)
	}
	if status.Initialized {
		t.Fatal("expected not initialized")
	}

	setup := storage.SetupRecord{
		ID:            "setup_1",
		Version:       "1.0.0",
		InitializedAt: time.Now().UTC().Format(time.RFC3339),
		UsageMode:     "personal",
	}
	if err := store.SaveSetup(ctx, setup); err != nil {
		t.Fatalf("SaveSetup: %v", err)
	}

	status, err = store.GetSetupStatus(ctx)
	if err != nil {
		t.Fatalf("GetSetupStatus after save: %v", err)
	}
	if !status.Initialized {
		t.Fatal("expected initialized")
	}
	if status.Version != "1.0.0" {
		t.Fatalf("expected version 1.0.0, got %s", status.Version)
	}
}

func testSessionLifecycle(t *testing.T, factory ProviderFactory) {
	store := factory(t)
	defer store.Close()
	ctx := context.Background()

	user := storage.UserRecord{
		ID: uuid.New().String(), UserName: "sess_" + uuid.New().String()[:8],
		PasswordHash: "h", PasswordSalt: "s", Roles: []string{"R_USER"}, Buttons: []string{"view"},
	}
	if _, err := store.CreateUser(ctx, user); err != nil {
		t.Fatal(err)
	}

	session := storage.SessionRecord{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	}
	if err := store.CreateSession(ctx, session); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	sessions, err := store.ListSessionsByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListSessionsByUser: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}

	if err := store.UpdateSessionLastSeen(ctx, session.ID); err != nil {
		t.Fatalf("UpdateSessionLastSeen: %v", err)
	}

	if err := store.RevokeSession(ctx, session.ID); err != nil {
		t.Fatalf("RevokeSession: %v", err)
	}

	sessions, _ = store.ListSessionsByUser(ctx, user.ID)
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions after revoke, got %d", len(sessions))
	}
}

func testAuditLog(t *testing.T, factory ProviderFactory) {
	store := factory(t)
	defer store.Close()
	ctx := context.Background()

	if err := store.CreateAuditLog(ctx, storage.AuditRecord{
		ID: uuid.New().String(), Action: "login", IPAddress: "127.0.0.1",
	}); err != nil {
		t.Fatalf("CreateAuditLog: %v", err)
	}

	logs, err := store.ListAuditLogs(ctx, "", "", 10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}

	total, err := store.CountAuditLogs(ctx, "", "")
	if err != nil {
		t.Fatalf("CountAuditLogs: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
}

func testPasskeyCredential(t *testing.T, factory ProviderFactory) {
	store := factory(t)
	defer store.Close()
	ctx := context.Background()

	user := storage.UserRecord{
		ID: uuid.New().String(), UserName: "pk_" + uuid.New().String()[:8],
		PasswordHash: "h", PasswordSalt: "s", Roles: []string{"R_USER"}, Buttons: []string{"view"},
	}
	if _, err := store.CreateUser(ctx, user); err != nil {
		t.Fatal(err)
	}

	cred := storage.PasskeyCredential{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		CredentialID: "cred_" + uuid.New().String(),
		PublicKey:    "pubkey_test",
		Transports:   `["usb","nfc"]`,
		Name:         "Test Key",
	}
	if err := store.CreatePasskeyCredential(ctx, cred); err != nil {
		t.Fatalf("CreatePasskeyCredential: %v", err)
	}

	got, err := store.GetPasskeyByCredentialID(ctx, cred.CredentialID)
	if err != nil {
		t.Fatalf("GetPasskeyByCredentialID: %v", err)
	}
	if got.Name != "Test Key" {
		t.Fatalf("name mismatch: %s", got.Name)
	}

	passkeys, err := store.ListPasskeysByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListPasskeysByUser: %v", err)
	}
	if len(passkeys) != 1 {
		t.Fatalf("expected 1 passkey, got %d", len(passkeys))
	}

	if err := store.UpdatePasskeyLastUsed(ctx, cred.ID, 42); err != nil {
		t.Fatalf("UpdatePasskeyLastUsed: %v", err)
	}

	if err := store.DeletePasskey(ctx, cred.ID, user.ID); err != nil {
		t.Fatalf("DeletePasskey: %v", err)
	}

	passkeys, _ = store.ListPasskeysByUser(ctx, user.ID)
	if len(passkeys) != 0 {
		t.Fatalf("expected 0 passkeys after delete, got %d", len(passkeys))
	}
}

func testChallengeLifecycle(t *testing.T, factory ProviderFactory) {
	store := factory(t)
	defer store.Close()
	ctx := context.Background()

	chal := storage.ChallengeRecord{
		ID:        uuid.New().String(),
		Challenge: "test_challenge_value",
		FlowType:  "passkey_login",
		ExpiresAt: time.Now().Add(5 * time.Minute).UTC(),
	}
	if err := store.CreateChallenge(ctx, chal); err != nil {
		t.Fatalf("CreateChallenge: %v", err)
	}

	got, err := store.GetChallenge(ctx, chal.ID)
	if err != nil {
		t.Fatalf("GetChallenge: %v", err)
	}
	if got.Challenge != "test_challenge_value" {
		t.Fatalf("challenge mismatch: %s", got.Challenge)
	}

	if err := store.DeleteChallenge(ctx, chal.ID); err != nil {
		t.Fatalf("DeleteChallenge: %v", err)
	}

	_, err = store.GetChallenge(ctx, chal.ID)
	if err == nil {
		t.Fatal("expected error for deleted challenge")
	}
}

func testAuthMethod(t *testing.T, factory ProviderFactory) {
	store := factory(t)
	defer store.Close()
	ctx := context.Background()

	user := storage.UserRecord{
		ID: uuid.New().String(), UserName: "am_" + uuid.New().String()[:8],
		PasswordHash: "h", PasswordSalt: "s", Roles: []string{"R_USER"}, Buttons: []string{"view"},
	}
	if _, err := store.CreateUser(ctx, user); err != nil {
		t.Fatal(err)
	}

	method := storage.AuthMethodRecord{
		ID:             uuid.New().String(),
		UserID:         user.ID,
		ProviderType:   "password",
		ProviderUserID: user.ID,
		Metadata:       "{}",
	}
	if err := store.CreateAuthMethod(ctx, method); err != nil {
		t.Fatalf("CreateAuthMethod: %v", err)
	}

	got, err := store.GetAuthMethodByProvider(ctx, "password", user.ID)
	if err != nil {
		t.Fatalf("GetAuthMethodByProvider: %v", err)
	}
	if got.UserID != user.ID {
		t.Fatalf("user ID mismatch: %s", got.UserID)
	}
}

// NewSQLiteTestProvider 创建一个临时 SQLite Provider 用于测试。
func NewSQLiteTestProvider(t *testing.T) storage.Provider {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.NewSQLiteProvider(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteProvider: %v", err)
	}
	if err := store.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return store
}
