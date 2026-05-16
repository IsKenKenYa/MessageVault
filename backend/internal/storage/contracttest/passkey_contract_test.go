package contracttest

import (
	"context"
	"testing"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/storage"
	"github.com/google/uuid"
)

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

	expired := storage.ChallengeRecord{
		ID:        uuid.New().String(),
		Challenge: "expired_challenge_value",
		FlowType:  "passkey_login",
		ExpiresAt: time.Now().Add(-1 * time.Minute).UTC(),
	}
	if err := store.CreateChallenge(ctx, expired); err != nil {
		t.Fatalf("CreateChallenge expired: %v", err)
	}
	if _, err := store.GetChallenge(ctx, expired.ID); err == nil {
		t.Fatal("expected error for expired challenge")
	}
}
