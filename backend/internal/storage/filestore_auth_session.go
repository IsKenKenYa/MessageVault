package storage

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

func (s *fileStore) CreateUser(ctx context.Context, user UserRecord) (UserRecord, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.snapshot.Users {
		if strings.EqualFold(existing.UserName, user.UserName) {
			return UserRecord{}, fmt.Errorf("username already exists")
		}
	}
	s.snapshot.Users[user.ID] = user
	return user, s.persist()
}

func (s *fileStore) FindUserByUserName(ctx context.Context, userName string) (UserRecord, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, user := range s.snapshot.Users {
		if strings.EqualFold(user.UserName, userName) {
			return user, nil
		}
	}
	return UserRecord{}, fmt.Errorf("user not found")
}

func (s *fileStore) GetUser(ctx context.Context, userID string) (UserRecord, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.snapshot.Users[userID]
	if !ok {
		return UserRecord{}, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *fileStore) SaveRefreshToken(ctx context.Context, token RefreshTokenRecord) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot.RefreshTokens[token.ID] = token
	return s.persist()
}

func (s *fileStore) ConsumeRefreshToken(ctx context.Context, tokenHash string) (RefreshTokenRecord, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, token := range s.snapshot.RefreshTokens {
		if token.TokenHash != tokenHash {
			continue
		}
		if !token.RevokedAt.IsZero() {
			return RefreshTokenRecord{}, fmt.Errorf("refresh token revoked")
		}
		if time.Now().After(token.ExpiresAt) {
			return RefreshTokenRecord{}, fmt.Errorf("refresh token expired")
		}
		token.RevokedAt = time.Now().UTC()
		s.snapshot.RefreshTokens[key] = token
		if err := s.persist(); err != nil {
			return RefreshTokenRecord{}, err
		}
		return token, nil
	}
	return RefreshTokenRecord{}, fmt.Errorf("refresh token not found")
}

func (s *fileStore) UpdateUserPasswordHash(ctx context.Context, userID, newHash, newSalt string) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.snapshot.Users[userID]
	if !ok {
		return fmt.Errorf("user not found")
	}
	user.PasswordHash = newHash
	user.PasswordSalt = newSalt
	user.UpdatedAt = time.Now().UTC()
	s.snapshot.Users[userID] = user
	return s.persist()
}

func (s *fileStore) FindAnyRefreshTokenByHash(_ context.Context, tokenHash string) (RefreshTokenRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, token := range s.snapshot.RefreshTokens {
		if token.TokenHash == tokenHash {
			return token, nil
		}
	}
	return RefreshTokenRecord{}, fmt.Errorf("refresh token not found")
}

func (s *fileStore) HasActiveRefreshTokenChild(_ context.Context, parentID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now().UTC()
	for _, token := range s.snapshot.RefreshTokens {
		if token.ParentID != parentID {
			continue
		}
		if token.RevokedAt.IsZero() && token.ExpiresAt.After(now) {
			return true, nil
		}
	}
	return false, nil
}

func (s *fileStore) RevokeRefreshTokenFamily(_ context.Context, tokenID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue := []string{tokenID}
	seen := map[string]struct{}{}
	now := time.Now().UTC()
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if _, ok := seen[current]; ok {
			continue
		}
		seen[current] = struct{}{}
		if token, ok := s.snapshot.RefreshTokens[current]; ok && token.RevokedAt.IsZero() {
			token.RevokedAt = now
			s.snapshot.RefreshTokens[current] = token
		}
		for id, token := range s.snapshot.RefreshTokens {
			if token.ParentID == current {
				queue = append(queue, id)
			}
		}
	}
	return s.persist()
}

func (s *fileStore) RevokeRefreshTokenByID(_ context.Context, tokenID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	token, ok := s.snapshot.RefreshTokens[tokenID]
	if !ok {
		return nil
	}
	if token.RevokedAt.IsZero() {
		token.RevokedAt = time.Now().UTC()
		s.snapshot.RefreshTokens[tokenID] = token
	}
	return s.persist()
}

func (s *fileStore) CreateSession(_ context.Context, rec SessionRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot.Sessions[rec.ID] = rec
	return s.persist()
}

func (s *fileStore) GetSession(_ context.Context, sessionID string) (SessionRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.snapshot.Sessions[sessionID]
	if !ok {
		return SessionRecord{}, fmt.Errorf("session not found")
	}
	return session, nil
}

func (s *fileStore) GetSessionByRefreshTokenID(_ context.Context, refreshTokenID string) (SessionRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, session := range s.snapshot.Sessions {
		if session.RefreshTokenID == refreshTokenID {
			return session, nil
		}
	}
	return SessionRecord{}, fmt.Errorf("session not found")
}

func (s *fileStore) ListSessionsByUser(_ context.Context, userID string) ([]SessionRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]SessionRecord, 0, len(s.snapshot.Sessions))
	for _, session := range s.snapshot.Sessions {
		if session.UserID != userID {
			continue
		}
		items = append(items, session)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].LastSeenAt.After(items[j].LastSeenAt) })
	return items, nil
}

func (s *fileStore) RevokeSession(_ context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.snapshot.Sessions[sessionID]
	if !ok {
		return nil
	}
	session.LastSeenAt = time.Now().UTC()
	delete(s.snapshot.Sessions, sessionID)
	if session.RefreshTokenID != "" {
		if token, ok := s.snapshot.RefreshTokens[session.RefreshTokenID]; ok && token.RevokedAt.IsZero() {
			token.RevokedAt = time.Now().UTC()
			s.snapshot.RefreshTokens[session.RefreshTokenID] = token
		}
	}
	return s.persist()
}

func (s *fileStore) RevokeOtherSessions(_ context.Context, userID, currentSessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	for id, session := range s.snapshot.Sessions {
		if session.UserID != userID || id == currentSessionID {
			continue
		}
		if session.RefreshTokenID != "" {
			if token, ok := s.snapshot.RefreshTokens[session.RefreshTokenID]; ok && token.RevokedAt.IsZero() {
				token.RevokedAt = now
				s.snapshot.RefreshTokens[session.RefreshTokenID] = token
			}
		}
		delete(s.snapshot.Sessions, id)
	}
	return s.persist()
}

func (s *fileStore) UpdateSessionLastSeen(_ context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.snapshot.Sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found")
	}
	session.LastSeenAt = time.Now().UTC()
	s.snapshot.Sessions[sessionID] = session
	return s.persist()
}

func (s *fileStore) UpdateSessionRefreshToken(_ context.Context, sessionID, refreshTokenID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.snapshot.Sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found")
	}
	session.RefreshTokenID = refreshTokenID
	session.LastSeenAt = time.Now().UTC()
	s.snapshot.Sessions[sessionID] = session
	return s.persist()
}
