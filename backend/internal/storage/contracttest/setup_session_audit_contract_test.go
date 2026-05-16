package contracttest

import (
	"context"
	"testing"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/storage"
	"github.com/google/uuid"
)

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

	refreshTokenID := uuid.New().String()
	if err := store.SaveRefreshToken(ctx, storage.RefreshTokenRecord{
		ID:        refreshTokenID,
		UserID:    user.ID,
		TokenHash: "hash_" + uuid.New().String(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}); err != nil {
		t.Fatalf("SaveRefreshToken: %v", err)
	}

	session := storage.SessionRecord{
		ID:             uuid.New().String(),
		UserID:         user.ID,
		RefreshTokenID: refreshTokenID,
		IPAddress:      "127.0.0.1",
		UserAgent:      "test-agent",
	}
	if err := store.CreateSession(ctx, session); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	byRefreshToken, err := store.GetSessionByRefreshTokenID(ctx, refreshTokenID)
	if err != nil {
		t.Fatalf("GetSessionByRefreshTokenID: %v", err)
	}
	if byRefreshToken.ID != session.ID {
		t.Fatalf("session id mismatch by refresh token: got %s", byRefreshToken.ID)
	}

	sessions, err := store.ListSessionsByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListSessionsByUser: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}

	otherRefreshTokenID := uuid.New().String()
	if err := store.SaveRefreshToken(ctx, storage.RefreshTokenRecord{
		ID:        otherRefreshTokenID,
		UserID:    user.ID,
		TokenHash: "hash_" + uuid.New().String(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}); err != nil {
		t.Fatalf("SaveRefreshToken other: %v", err)
	}
	otherSession := storage.SessionRecord{
		ID:             uuid.New().String(),
		UserID:         user.ID,
		RefreshTokenID: otherRefreshTokenID,
	}
	if err := store.CreateSession(ctx, otherSession); err != nil {
		t.Fatalf("CreateSession other: %v", err)
	}

	if err := store.RevokeOtherSessions(ctx, user.ID, session.ID); err != nil {
		t.Fatalf("RevokeOtherSessions: %v", err)
	}
	sessions, err = store.ListSessionsByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListSessionsByUser after revoke others: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != session.ID {
		t.Fatalf("expected current session to remain after revoke others, got %+v", sessions)
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
