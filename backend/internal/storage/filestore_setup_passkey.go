package storage

import (
	"context"
	"fmt"
	"sort"
	"time"
)

func (s *fileStore) GetSetupStatus(ctx context.Context) (SetupStatus, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.snapshot.Setup != nil {
		return SetupStatus{Initialized: true, Version: s.snapshot.Setup.Version, DatabaseType: "sqlite"}, nil
	}
	hasAdmin, _ := s.hasAdminUserLocked()
	return SetupStatus{Initialized: hasAdmin, DatabaseType: "sqlite"}, nil
}

func (s *fileStore) SaveSetup(ctx context.Context, setup SetupRecord) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot.Setup = &setup
	return s.persist()
}

func (s *fileStore) HasAdminUser(ctx context.Context) (bool, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hasAdminUserLocked()
}

func (s *fileStore) hasAdminUserLocked() (bool, error) {
	for _, user := range s.snapshot.Users {
		for _, role := range user.Roles {
			if role == "R_ADMIN" || role == "R_SUPER" {
				return true, nil
			}
		}
	}
	return false, nil
}

func (s *fileStore) CreateAuditLog(_ context.Context, rec AuditRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot.AuditLogs[rec.ID] = rec
	return s.persist()
}

func (s *fileStore) ListAuditLogs(_ context.Context, userID, action string, limit, offset int) ([]AuditRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]AuditRecord, 0, len(s.snapshot.AuditLogs))
	for _, rec := range s.snapshot.AuditLogs {
		if userID != "" && rec.UserID != userID {
			continue
		}
		if action != "" && rec.Action != action {
			continue
		}
		items = append(items, rec)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	if offset > 0 && offset < len(items) {
		items = items[offset:]
	} else if offset >= len(items) {
		return nil, nil
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (s *fileStore) CountAuditLogs(ctx context.Context, userID, action string) (int, error) {
	items, err := s.ListAuditLogs(ctx, userID, action, 0, 0)
	if err != nil {
		return 0, err
	}
	return len(items), nil
}

func (s *fileStore) CreatePasskeyCredential(_ context.Context, rec PasskeyCredential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot.Passkeys[rec.ID] = rec
	return s.persist()
}

func (s *fileStore) GetPasskeyByCredentialID(_ context.Context, credentialID string) (PasskeyCredential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, passkey := range s.snapshot.Passkeys {
		if passkey.CredentialID == credentialID {
			return passkey, nil
		}
	}
	return PasskeyCredential{}, fmt.Errorf("passkey not found")
}

func (s *fileStore) ListPasskeysByUser(_ context.Context, userID string) ([]PasskeyCredential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]PasskeyCredential, 0, len(s.snapshot.Passkeys))
	for _, passkey := range s.snapshot.Passkeys {
		if passkey.UserID != userID {
			continue
		}
		items = append(items, passkey)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items, nil
}

func (s *fileStore) UpdatePasskeyLastUsed(_ context.Context, id string, signCount uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	passkey, ok := s.snapshot.Passkeys[id]
	if !ok {
		return fmt.Errorf("passkey not found")
	}
	passkey.SignCount = signCount
	passkey.LastUsedAt = time.Now().UTC()
	s.snapshot.Passkeys[id] = passkey
	return s.persist()
}

func (s *fileStore) DeletePasskey(_ context.Context, id, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	passkey, ok := s.snapshot.Passkeys[id]
	if !ok || passkey.UserID != userID {
		return nil
	}
	delete(s.snapshot.Passkeys, id)
	return s.persist()
}

func (s *fileStore) CreateChallenge(_ context.Context, rec ChallengeRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = time.Now().UTC()
	}
	s.snapshot.Challenges[rec.ID] = rec
	return s.persist()
}

func (s *fileStore) GetChallenge(_ context.Context, id string) (ChallengeRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	challenge, ok := s.snapshot.Challenges[id]
	if !ok || time.Now().UTC().After(challenge.ExpiresAt) {
		return ChallengeRecord{}, fmt.Errorf("challenge not found")
	}
	return challenge, nil
}

func (s *fileStore) DeleteChallenge(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.snapshot.Challenges, id)
	return s.persist()
}

func (s *fileStore) CreateAuthMethod(_ context.Context, rec AuthMethodRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.snapshot.AuthMethods {
		if existing.ProviderType == rec.ProviderType && existing.ProviderUserID == rec.ProviderUserID {
			return fmt.Errorf("auth method already exists")
		}
	}
	s.snapshot.AuthMethods[rec.ID] = rec
	return s.persist()
}

func (s *fileStore) GetAuthMethodByProvider(_ context.Context, providerType, providerUserID string) (AuthMethodRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, rec := range s.snapshot.AuthMethods {
		if rec.ProviderType == providerType && rec.ProviderUserID == providerUserID {
			return rec, nil
		}
	}
	return AuthMethodRecord{}, fmt.Errorf("auth method not found")
}
