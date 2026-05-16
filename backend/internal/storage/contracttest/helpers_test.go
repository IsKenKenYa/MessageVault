package contracttest

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

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
