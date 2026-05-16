package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	sqlc "github.com/IsKenKenYa/Commory/backend/internal/storage/sqlc/gen"
)

func (s *sqliteProvider) CreateUser(ctx context.Context, user UserRecord) (UserRecord, error) {
	roles, _ := json.Marshal(user.Roles)
	buttons, _ := json.Marshal(user.Buttons)
	err := s.q.CreateUser(ctx, &sqlc.CreateUserParams{
		ID:           user.ID,
		UserName:     user.UserName,
		Email:        sql.NullString{String: user.Email, Valid: user.Email != ""},
		PasswordHash: user.PasswordHash,
		PasswordSalt: user.PasswordSalt,
		Roles:        string(roles),
		Buttons:      string(buttons),
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return UserRecord{}, fmt.Errorf("username already exists")
		}
		return UserRecord{}, err
	}
	return s.GetUser(ctx, user.ID)
}

func (s *sqliteProvider) FindUserByUserName(ctx context.Context, userName string) (UserRecord, error) {
	row, err := s.q.FindUserByUserName(ctx, userName)
	if err != nil {
		return UserRecord{}, fmt.Errorf("user not found")
	}
	return s.rowToUser(row), nil
}

func (s *sqliteProvider) GetUser(ctx context.Context, userID string) (UserRecord, error) {
	row, err := s.q.GetUser(ctx, userID)
	if err != nil {
		return UserRecord{}, fmt.Errorf("user not found")
	}
	return s.rowToUser(row), nil
}

func (s *sqliteProvider) SaveRefreshToken(ctx context.Context, token RefreshTokenRecord) error {
	return s.q.SaveRefreshToken(ctx, &sqlc.SaveRefreshTokenParams{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		ParentID:  sql.NullString{String: token.ParentID, Valid: token.ParentID != ""},
		ExpiresAt: token.ExpiresAt,
	})
}

func (s *sqliteProvider) ConsumeRefreshToken(ctx context.Context, tokenHash string) (RefreshTokenRecord, error) {
	row, err := s.q.ConsumeRefreshToken(ctx, tokenHash)
	if err != nil {
		return RefreshTokenRecord{}, fmt.Errorf("refresh token not found or expired")
	}
	return RefreshTokenRecord{
		ID:        row.ID,
		UserID:    row.UserID,
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
		RevokedAt: row.RevokedAt.Time,
	}, nil
}

func (s *sqliteProvider) UpdateUserPasswordHash(ctx context.Context, userID, newHash, newSalt string) error {
	return s.q.UpdateUserPasswordHash(ctx, &sqlc.UpdateUserPasswordHashParams{
		PasswordHash: newHash,
		PasswordSalt: newSalt,
		ID:           userID,
	})
}

func (s *sqliteProvider) FindAnyRefreshTokenByHash(ctx context.Context, tokenHash string) (RefreshTokenRecord, error) {
	row, err := s.q.FindAnyRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return RefreshTokenRecord{}, err
	}
	return s.rowToRefreshToken(row), nil
}

func (s *sqliteProvider) HasActiveRefreshTokenChild(ctx context.Context, parentID string) (bool, error) {
	return s.q.HasActiveRefreshTokenChild(ctx, sql.NullString{String: parentID, Valid: parentID != ""})
}

func (s *sqliteProvider) RevokeRefreshTokenFamily(ctx context.Context, tokenID string) error {
	return s.q.RevokeRefreshTokenFamily(ctx, tokenID)
}

func (s *sqliteProvider) RevokeRefreshTokenByID(ctx context.Context, tokenID string) error {
	return s.q.RevokeRefreshTokenByID(ctx, tokenID)
}

func (s *sqliteProvider) CreateSession(ctx context.Context, rec SessionRecord) error {
	return s.q.CreateSession(ctx, &sqlc.CreateSessionParams{
		ID:             rec.ID,
		UserID:         rec.UserID,
		RefreshTokenID: sql.NullString{String: rec.RefreshTokenID, Valid: rec.RefreshTokenID != ""},
		DeviceName:     sql.NullString{String: rec.DeviceName, Valid: rec.DeviceName != ""},
		DeviceType:     sql.NullString{String: rec.DeviceType, Valid: rec.DeviceType != ""},
		IpAddress:      sql.NullString{String: rec.IPAddress, Valid: rec.IPAddress != ""},
		UserAgent:      sql.NullString{String: rec.UserAgent, Valid: rec.UserAgent != ""},
	})
}

func (s *sqliteProvider) GetSession(ctx context.Context, sessionID string) (SessionRecord, error) {
	row, err := s.q.GetSession(ctx, sessionID)
	if err != nil {
		return SessionRecord{}, err
	}
	return s.rowToSession(row), nil
}

func (s *sqliteProvider) GetSessionByRefreshTokenID(ctx context.Context, refreshTokenID string) (SessionRecord, error) {
	row, err := s.q.GetSessionByRefreshTokenID(ctx, sql.NullString{String: refreshTokenID, Valid: refreshTokenID != ""})
	if err != nil {
		return SessionRecord{}, err
	}
	return s.rowToSession(row), nil
}

func (s *sqliteProvider) ListSessionsByUser(ctx context.Context, userID string) ([]SessionRecord, error) {
	rows, err := s.q.ListSessionsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]SessionRecord, 0, len(rows))
	for _, r := range rows {
		items = append(items, s.rowToSession(r))
	}
	return items, nil
}

func (s *sqliteProvider) RevokeSession(ctx context.Context, sessionID string) error {
	return s.q.RevokeSession(ctx, sessionID)
}

func (s *sqliteProvider) RevokeOtherSessions(ctx context.Context, userID, currentSessionID string) error {
	return s.q.RevokeOtherSessions(ctx, &sqlc.RevokeOtherSessionsParams{
		UserID: userID,
		ID:     currentSessionID,
	})
}

func (s *sqliteProvider) UpdateSessionLastSeen(ctx context.Context, sessionID string) error {
	return s.q.UpdateSessionLastSeen(ctx, sessionID)
}

func (s *sqliteProvider) UpdateSessionRefreshToken(ctx context.Context, sessionID, refreshTokenID string) error {
	return s.q.UpdateSessionRefreshToken(ctx, &sqlc.UpdateSessionRefreshTokenParams{
		RefreshTokenID: sql.NullString{String: refreshTokenID, Valid: refreshTokenID != ""},
		ID:             sessionID,
	})
}
