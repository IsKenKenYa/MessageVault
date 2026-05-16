package storage

import (
	"context"
	"database/sql"
	"time"

	sqlc "github.com/IsKenKenYa/Commory/backend/internal/storage/sqlc/gen"
)

func (s *sqliteProvider) GetSetupStatus(ctx context.Context) (SetupStatus, error) {
	row, err := s.q.GetSetupStatus(ctx)
	if err != nil {
		hasAdmin, _ := s.q.HasAdminUser(ctx)
		return SetupStatus{Initialized: hasAdmin, DatabaseType: "sqlite"}, nil
	}
	return SetupStatus{
		Initialized:  true,
		Version:      row.Version,
		DatabaseType: "sqlite",
	}, nil
}

func (s *sqliteProvider) SaveSetup(ctx context.Context, setup SetupRecord) error {
	ts, _ := time.Parse(time.RFC3339, setup.InitializedAt)
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	return s.q.SaveSetup(ctx, &sqlc.SaveSetupParams{
		ID:            setup.ID,
		Version:       setup.Version,
		InitializedAt: ts,
		UsageMode:     sql.NullString{String: setup.UsageMode, Valid: setup.UsageMode != ""},
	})
}

func (s *sqliteProvider) HasAdminUser(ctx context.Context) (bool, error) {
	return s.q.HasAdminUser(ctx)
}

func (s *sqliteProvider) CreateAuditLog(ctx context.Context, rec AuditRecord) error {
	return s.q.CreateAuditLog(ctx, &sqlc.CreateAuditLogParams{
		ID:        rec.ID,
		UserID:    sql.NullString{String: rec.UserID, Valid: rec.UserID != ""},
		Action:    rec.Action,
		IpAddress: sql.NullString{String: rec.IPAddress, Valid: rec.IPAddress != ""},
		UserAgent: sql.NullString{String: rec.UserAgent, Valid: rec.UserAgent != ""},
		Detail:    sql.NullString{String: rec.Detail, Valid: rec.Detail != ""},
	})
}

func (s *sqliteProvider) ListAuditLogs(ctx context.Context, userID, action string, limit, offset int) ([]AuditRecord, error) {
	rows, err := s.q.GetAuditLogs(ctx, &sqlc.GetAuditLogsParams{
		UserID: userID,
		Action: action,
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		return nil, err
	}
	items := make([]AuditRecord, 0, len(rows))
	for _, r := range rows {
		items = append(items, AuditRecord{
			ID:        r.ID,
			UserID:    r.UserID.String,
			Action:    r.Action,
			IPAddress: r.IpAddress.String,
			UserAgent: r.UserAgent.String,
			Detail:    r.Detail.String,
			CreatedAt: r.CreatedAt,
		})
	}
	return items, nil
}

func (s *sqliteProvider) CountAuditLogs(ctx context.Context, userID, action string) (int, error) {
	count, err := s.q.CountAuditLogs(ctx, &sqlc.CountAuditLogsParams{
		UserID: userID,
		Action: action,
	})
	return int(count), err
}

func (s *sqliteProvider) CreatePasskeyCredential(ctx context.Context, rec PasskeyCredential) error {
	return s.q.CreatePasskeyCredential(ctx, &sqlc.CreatePasskeyCredentialParams{
		ID:              rec.ID,
		UserID:          rec.UserID,
		CredentialID:    rec.CredentialID,
		PublicKey:       rec.PublicKey,
		AttestationType: rec.AttestationType,
		Aaguid:          rec.AAGUID,
		SignCount:       int64(rec.SignCount),
		Transports:      rec.Transports,
		Name:            rec.Name,
	})
}

func (s *sqliteProvider) GetPasskeyByCredentialID(ctx context.Context, credentialID string) (PasskeyCredential, error) {
	row, err := s.q.GetPasskeyByCredentialID(ctx, credentialID)
	if err != nil {
		return PasskeyCredential{}, err
	}
	return s.rowToPasskey(row), nil
}

func (s *sqliteProvider) ListPasskeysByUser(ctx context.Context, userID string) ([]PasskeyCredential, error) {
	rows, err := s.q.ListPasskeysByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]PasskeyCredential, 0, len(rows))
	for _, r := range rows {
		items = append(items, s.rowToPasskey(r))
	}
	return items, nil
}

func (s *sqliteProvider) UpdatePasskeyLastUsed(ctx context.Context, id string, signCount uint32) error {
	return s.q.UpdatePasskeyLastUsed(ctx, &sqlc.UpdatePasskeyLastUsedParams{
		SignCount: int64(signCount),
		ID:        id,
	})
}

func (s *sqliteProvider) DeletePasskey(ctx context.Context, id, userID string) error {
	return s.q.DeletePasskey(ctx, &sqlc.DeletePasskeyParams{ID: id, UserID: userID})
}

func (s *sqliteProvider) CreateChallenge(ctx context.Context, rec ChallengeRecord) error {
	return s.q.CreateChallenge(ctx, &sqlc.CreateChallengeParams{
		ID:        rec.ID,
		Challenge: rec.Challenge,
		UserID:    sql.NullString{String: rec.UserID, Valid: rec.UserID != ""},
		FlowType:  rec.FlowType,
		ExpiresAt: rec.ExpiresAt,
	})
}

func (s *sqliteProvider) GetChallenge(ctx context.Context, id string) (ChallengeRecord, error) {
	row, err := s.q.GetChallenge(ctx, id)
	if err != nil {
		return ChallengeRecord{}, err
	}
	return ChallengeRecord{
		ID:        row.ID,
		Challenge: row.Challenge,
		UserID:    row.UserID.String,
		FlowType:  row.FlowType,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (s *sqliteProvider) DeleteChallenge(ctx context.Context, id string) error {
	return s.q.DeleteChallenge(ctx, id)
}

func (s *sqliteProvider) CreateAuthMethod(ctx context.Context, rec AuthMethodRecord) error {
	return s.q.CreateAuthMethod(ctx, &sqlc.CreateAuthMethodParams{
		ID:             rec.ID,
		UserID:         rec.UserID,
		ProviderType:   rec.ProviderType,
		ProviderUserID: rec.ProviderUserID,
		Metadata:       rec.Metadata,
	})
}

func (s *sqliteProvider) GetAuthMethodByProvider(ctx context.Context, providerType, providerUserID string) (AuthMethodRecord, error) {
	row, err := s.q.GetAuthMethodByProvider(ctx, &sqlc.GetAuthMethodByProviderParams{
		ProviderType:   providerType,
		ProviderUserID: providerUserID,
	})
	if err != nil {
		return AuthMethodRecord{}, err
	}
	return AuthMethodRecord{
		ID:             row.ID,
		UserID:         row.UserID,
		ProviderType:   row.ProviderType,
		ProviderUserID: row.ProviderUserID,
		Metadata:       row.Metadata,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}, nil
}
