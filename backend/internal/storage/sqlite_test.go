package storage

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/importers"
	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
)

func TestSQLiteImportAndQuery(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "commory.sqlite")
	store, err := NewSQLiteProvider(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(ctx); err != nil {
		t.Fatal(err)
	}

	rootSchema := repoPath("msglayer", "schema", "v0.1", "root.schema.json")
	validator, err := msglayer.NewValidator(rootSchema)
	if err != nil {
		t.Fatal(err)
	}
	sample := repoPath("msglayer", "examples", "export.sms-call.json")
	importer := importers.JSONImporter{}
	export, raw, err := importer.Import(sample)
	if err != nil {
		t.Fatal(err)
	}
	if err := validator.ValidateBytes(raw); err != nil {
		t.Fatal(err)
	}
	importID, err := store.Import(ctx, "test-user", sample, export, raw)
	if err != nil {
		t.Fatal(err)
	}
	if importID == "" {
		t.Fatal("expected import id")
	}

	items, err := store.Search(ctx, msglayer.SearchParams{UserID: "test-user", Keyword: "验证码", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatal("expected keyword search hits")
	}

	thread, err := store.GetThread(ctx, "test-user", "thread_42")
	if err != nil {
		t.Fatal(err)
	}
	if len(thread) == 0 {
		t.Fatal("expected thread reconstruction results")
	}

	rawExport, err := store.ExportImport(ctx, "test-user", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(rawExport) == 0 {
		t.Fatal("expected exported raw json")
	}
}

func TestPostgresProviderSkipsWithoutDSN(t *testing.T) {
	if os.Getenv("COMMORY_TEST_POSTGRES_DSN") == "" {
		t.Skip("COMMORY_TEST_POSTGRES_DSN not set")
	}
}

func TestSQLiteAuditLogBlankUserFilterIncludesSystemEntries(t *testing.T) {
	ctx := context.Background()
	store, err := NewSQLiteProvider(filepath.Join(t.TempDir(), "audit.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(ctx); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	if _, err := store.CreateUser(ctx, UserRecord{
		ID:           "user_1",
		UserName:     "audit-user",
		PasswordHash: "hash",
		PasswordSalt: "salt",
		Roles:        []string{"R_USER"},
		Buttons:      []string{"view"},
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	for _, rec := range []AuditRecord{
		{ID: "audit_sys", Action: "system_boot", CreatedAt: now},
		{ID: "audit_user", UserID: "user_1", Action: "login", CreatedAt: now.Add(time.Second)},
	} {
		if err := store.CreateAuditLog(ctx, rec); err != nil {
			t.Fatalf("CreateAuditLog(%s): %v", rec.ID, err)
		}
	}

	logs, err := store.ListAuditLogs(ctx, "", "", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 audit logs, got %d", len(logs))
	}
	count, err := store.CountAuditLogs(ctx, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected audit log count 2, got %d", count)
	}
}

func repoPath(parts ...string) string {
	_, file, _, _ := runtime.Caller(0)
	base := filepath.Join(filepath.Dir(file), "..", "..", "..")
	all := append([]string{base}, parts...)
	return filepath.Join(all...)
}
