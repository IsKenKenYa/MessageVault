package auth

import (
	"context"
	"log"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

const sessionLastSeenThrottle = 5 * time.Minute

func (s *Service) AuthenticateRequest(ctx context.Context, authHeader string) (AccessTokenClaims, error) {
	claims, err := s.ParseAccessTokenClaims(authHeader)
	if err != nil {
		return AccessTokenClaims{}, ErrUnauthorized
	}
	if claims.SessionID == "" {
		return AccessTokenClaims{}, ErrUnauthorized
	}
	session, err := s.GetSession(ctx, claims.SessionID)
	if err != nil {
		return AccessTokenClaims{}, ErrUnauthorized
	}
	if session.UserID != claims.UserID {
		return AccessTokenClaims{}, ErrUnauthorized
	}
	if shouldUpdateSessionLastSeen(session, time.Now().UTC()) {
		if err := s.UpdateSessionLastSeen(ctx, session.ID); err != nil {
			log.Printf("[auth] update session last_seen_at for %s: %v", session.ID, err)
		}
	}
	return claims, nil
}

func (s *Service) GetSession(ctx context.Context, sessionID string) (storage.SessionRecord, error) {
	return s.store.GetSession(ctx, sessionID)
}

func (s *Service) UpdateSessionLastSeen(ctx context.Context, sessionID string) error {
	return s.store.UpdateSessionLastSeen(ctx, sessionID)
}

func shouldUpdateSessionLastSeen(session storage.SessionRecord, now time.Time) bool {
	if session.LastSeenAt.IsZero() {
		return true
	}
	return now.Sub(session.LastSeenAt) >= sessionLastSeenThrottle
}
