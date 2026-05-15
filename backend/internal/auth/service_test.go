package auth

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestHashPasswordReturnsPBKDF2Prefix(t *testing.T) {
	salt, hash, err := hashPassword("testpassword")
	if err != nil {
		t.Fatalf("hashPassword error: %v", err)
	}
	if len(salt) == 0 {
		t.Fatal("salt should not be empty")
	}
	prefix := "pbkdf2$"
	if !strings.HasPrefix(hash, prefix) {
		t.Fatalf("hash should start with %q, got %q", prefix, hash)
	}
}

func TestVerifyPasswordNewPBKDF2Format(t *testing.T) {
	salt, hash, err := hashPassword("mypassword123")
	if err != nil {
		t.Fatalf("hashPassword error: %v", err)
	}
	if !verifyPassword("mypassword123", salt, hash) {
		t.Fatal("verifyPassword should succeed with correct password and PBKDF2 format")
	}
	if verifyPassword("wrongpassword", salt, hash) {
		t.Fatal("verifyPassword should fail with wrong password")
	}
}

func TestVerifyPasswordLegacyFormatNoPrefix(t *testing.T) {
	salt := "abcd1234ef567890"
	sum := derivePasswordLegacy("legacytest", salt)
	legacyHash := hex.EncodeToString(sum[:])
	if !verifyPassword("legacytest", salt, legacyHash) {
		t.Fatal("verifyPassword should succeed with legacy format (no prefix)")
	}
	if verifyPassword("wrongpassword", salt, legacyHash) {
		t.Fatal("verifyPassword should fail with wrong password for legacy format")
	}
}

func TestVerifyPasswordLegacyFormatWithSHA256Prefix(t *testing.T) {
	salt := "abcd1234ef567890"
	sum := derivePasswordLegacy("testpw", salt)
	legacyHash := "sha256$" + hex.EncodeToString(sum[:])
	if !verifyPassword("testpw", salt, legacyHash) {
		t.Fatal("verifyPassword should succeed with sha256$ prefix format")
	}
	if verifyPassword("wrongpw", salt, legacyHash) {
		t.Fatal("verifyPassword should fail with wrong password for sha256$ prefix")
	}
}

func TestNeedsRehash(t *testing.T) {
	if NeedsRehash("pbkdf2$abc123") {
		t.Fatal("NeedsRehash should return false for pbkdf2$ prefix")
	}
	if !NeedsRehash("abc123") {
		t.Fatal("NeedsRehash should return true for no prefix")
	}
	if !NeedsRehash("sha256$abc123") {
		t.Fatal("NeedsRehash should return true for sha256$ prefix")
	}
}
