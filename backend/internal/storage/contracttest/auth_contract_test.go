package contracttest

import (
	"context"
	"testing"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/storage"
	"github.com/google/uuid"
)

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

	_, err = store.ConsumeRefreshToken(ctx, token.TokenHash)
	if err == nil {
		t.Fatal("expected error for already consumed token")
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
