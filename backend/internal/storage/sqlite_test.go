package storage

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

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
	importID, err := store.Import(ctx, sample, export, raw)
	if err != nil {
		t.Fatal(err)
	}
	if importID == "" {
		t.Fatal("expected import id")
	}

	items, err := store.Search(ctx, msglayer.SearchParams{Keyword: "验证码", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatal("expected keyword search hits")
	}

	thread, err := store.GetThread(ctx, "thread_42")
	if err != nil {
		t.Fatal(err)
	}
	if len(thread) == 0 {
		t.Fatal("expected thread reconstruction results")
	}

	rawExport, err := store.ExportImport(ctx, "")
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

func repoPath(parts ...string) string {
	_, file, _, _ := runtime.Caller(0)
	base := filepath.Join(filepath.Dir(file), "..", "..", "..")
	all := append([]string{base}, parts...)
	return filepath.Join(all...)
}
