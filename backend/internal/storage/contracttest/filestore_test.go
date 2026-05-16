package contracttest

import (
	"path/filepath"
	"testing"

	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

func TestFileStoreContract(t *testing.T) {
	RunContractTests(t, func(t *testing.T) storage.Provider {
		t.Helper()
		store, err := storage.NewFileStoreProvider(filepath.Join(t.TempDir(), "commory.json"))
		if err != nil {
			t.Fatalf("NewFileStoreProvider: %v", err)
		}
		return store
	})
}
