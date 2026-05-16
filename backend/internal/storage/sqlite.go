package storage

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
	sqlc "github.com/IsKenKenYa/Commory/backend/internal/storage/sqlc/gen"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/sqlite/0001_initial.up.sql
var sqliteMigration001 string

type sqliteProvider struct {
	db  *sql.DB
	q   *sqlc.Queries
	dsn string
}

func NewSQLiteProvider(dsn string) (Provider, error) {
	db, err := sql.Open("sqlite3", dsn+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	return &sqliteProvider{
		db:  db,
		q:   sqlc.New(db),
		dsn: dsn,
	}, nil
}

func (s *sqliteProvider) Name() string { return "sqlite" }

func (s *sqliteProvider) Close() error { return s.db.Close() }

func (s *sqliteProvider) Init(ctx context.Context) error {
	// 运行迁移
	if _, err := s.db.ExecContext(ctx, sqliteMigration001); err != nil {
		return fmt.Errorf("run migration 0001: %w", err)
	}
	return nil
}

// ==================== Import ====================

func (s *sqliteProvider) Import(ctx context.Context, userID, sourcePath string, export msglayer.RootExport, raw []byte) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	qtx := s.q.WithTx(tx)

	importID := fmt.Sprintf("import_%d", time.Now().UnixNano())
	if err := qtx.CreateImport(ctx, &sqlc.CreateImportParams{
		ID:            importID,
		UserID:        userID,
		SchemaVersion: export.Version,
		IsDelta:       false,
		SourcePath:    sql.NullString{String: sourcePath, Valid: sourcePath != ""},
		EventCount:    int64(len(export.Events)),
		IdentityCount: int64(len(export.Identities)),
		RawJson:       string(raw),
	}); err != nil {
		return "", fmt.Errorf("create import: %w", err)
	}

	for _, identity := range export.Identities {
		phones, _ := json.Marshal(identity.Phones)
		emails, _ := json.Marshal(identity.Emails)
		labels, _ := json.Marshal(identity.Labels)
		meta, _ := json.Marshal(identity.Meta)
		var avatar sql.NullString
		if identity.Avatar != nil {
			avatar = sql.NullString{String: *identity.Avatar, Valid: true}
		}
		if err := qtx.CreateIdentity(ctx, &sqlc.CreateIdentityParams{
			ID:          identity.ID,
			UserID:      userID,
			Type:        identity.Type,
			DisplayName: identity.DisplayName,
			Phones:      string(phones),
			Emails:      string(emails),
			Avatar:      avatar,
			Labels:      string(labels),
			Meta:        string(meta),
		}); err != nil {
			return "", fmt.Errorf("create identity %s: %w", identity.ID, err)
		}
	}

	for _, event := range export.Events {
		contentJSON, _ := json.Marshal(event.Content)
		metaJSON, _ := json.Marshal(event.Meta)
		ts, _ := time.Parse(time.RFC3339, event.Timestamp)
		summary := summarizeEvent(event)

		if err := qtx.CreateEvent(ctx, &sqlc.CreateEventParams{
			ID:             event.ID,
			UserID:         userID,
			ImportID:       sql.NullString{String: importID, Valid: true},
			Type:           event.Type,
			Timestamp:      ts,
			Direction:      event.Direction,
			ContentSummary: sql.NullString{String: summary, Valid: summary != ""},
			Content:        string(contentJSON),
			Meta:           string(metaJSON),
		}); err != nil {
			return "", fmt.Errorf("create event %s: %w", event.ID, err)
		}

		for _, participant := range event.Participants {
			if err := qtx.InsertEventParticipant(ctx, &sqlc.InsertEventParticipantParams{
				EventID:    event.ID,
				IdentityID: participant,
			}); err != nil {
				return "", fmt.Errorf("insert participant %s: %w", participant, err)
			}
		}

		for _, rel := range event.Relations {
			if err := qtx.CreateRelation(ctx, &sqlc.CreateRelationParams{
				ID:      fmt.Sprintf("rel_%s_%s_%d", event.ID, rel.Type, time.Now().UnixNano()),
				EventID: event.ID,
				Type:    rel.Type,
				Target:  rel.Target,
			}); err != nil {
				return "", fmt.Errorf("create relation: %w", err)
			}
		}
	}

	return importID, tx.Commit()
}

func (s *sqliteProvider) ListImports(ctx context.Context, userID string) ([]ImportSummary, error) {
	rows, err := s.q.ListImportsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]ImportSummary, 0, len(rows))
	for _, r := range rows {
		items = append(items, ImportSummary{
			ID:            r.ID,
			UserID:        r.UserID,
			SchemaVersion: r.SchemaVersion,
			ImportedAt:    r.ImportedAt,
			SourcePath:    r.SourcePath.String,
			EventCount:    int(r.EventCount),
			IdentityCount: int(r.IdentityCount),
		})
	}
	return items, nil
}

func (s *sqliteProvider) ExportImport(ctx context.Context, userID, importID string) ([]byte, error) {
	if importID == "" {
		var err error
		importID, err = s.q.LatestImportID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("no imports found")
		}
	}
	raw, err := s.q.ExportImport(ctx, &sqlc.ExportImportParams{ID: importID, UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("import not found: %s", importID)
	}
	return []byte(raw), nil
}

func (s *sqliteProvider) LatestImportID(ctx context.Context, userID string) (string, error) {
	id, err := s.q.LatestImportID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("no imports found")
	}
	return id, nil
}

// ==================== Events ====================

func (s *sqliteProvider) GetEvent(ctx context.Context, userID, id string) (msglayer.TimelineItem, error) {
	row, err := s.q.GetEvent(ctx, &sqlc.GetEventParams{ID: id, UserID: userID})
	if err != nil {
		return msglayer.TimelineItem{}, fmt.Errorf("event not found: %s", id)
	}
	return s.eventToTimelineItem(ctx, row)
}

func (s *sqliteProvider) ListEvents(ctx context.Context, params msglayer.SearchParams) ([]msglayer.TimelineItem, error) {
	typeFilter := ""
	if params.Type != "" {
		typeFilter = params.Type
	}
	limit := int64(50)
	if params.Limit > 0 {
		limit = int64(params.Limit)
	}
	rows, err := s.q.ListEvents(ctx, &sqlc.ListEventsParams{
		UserID: params.UserID,
		Type:   typeFilter,
		Limit:  limit,
		Offset: int64(params.Offset),
	})
	if err != nil {
		return nil, err
	}
	return s.eventsToTimelineItems(ctx, rows)
}

func (s *sqliteProvider) Search(ctx context.Context, params msglayer.SearchParams) ([]msglayer.TimelineItem, error) {
	if params.Keyword == "" {
		return nil, nil
	}
	typeFilter := ""
	if params.Type != "" {
		typeFilter = params.Type
	}
	limit := int64(50)
	if params.Limit > 0 {
		limit = int64(params.Limit)
	}
	rows, err := s.q.SearchEvents(ctx, &sqlc.SearchEventsParams{
		UserID:     params.UserID,
		SearchTerm: sql.NullString{String: params.Keyword, Valid: true},
		Type:       typeFilter,
		Limit:      limit,
		Offset:     int64(params.Offset),
	})
	if err != nil {
		return nil, err
	}
	return s.eventsToTimelineItems(ctx, rows)
}

func (s *sqliteProvider) Timeline(ctx context.Context, params msglayer.SearchParams) ([]msglayer.TimelineItem, error) {
	typeFilter := ""
	if params.Type != "" {
		typeFilter = params.Type
	}
	limit := int64(50)
	if params.Limit > 0 {
		limit = int64(params.Limit)
	}

	if params.Participant != "" {
		rows, err := s.q.TimelineEventsByParticipant(ctx, &sqlc.TimelineEventsByParticipantParams{
			UserID:      params.UserID,
			Type:        typeFilter,
			Participant: sql.NullString{String: params.Participant, Valid: true},
			Limit:       limit,
			Offset:      int64(params.Offset),
		})
		if err != nil {
			return nil, err
		}
		return s.eventsToTimelineItems(ctx, rows)
	}

	rows, err := s.q.TimelineEvents(ctx, &sqlc.TimelineEventsParams{
		UserID: params.UserID,
		Type:   typeFilter,
		Limit:  limit,
		Offset: int64(params.Offset),
	})
	if err != nil {
		return nil, err
	}
	return s.eventsToTimelineItems(ctx, rows)
}

// ==================== Identities ====================

func (s *sqliteProvider) ListIdentities(ctx context.Context, userID string) ([]msglayer.Identity, error) {
	rows, err := s.q.ListIdentities(ctx, &sqlc.ListIdentitiesParams{UserID: userID, Type: ""})
	if err != nil {
		return nil, err
	}
	items := make([]msglayer.Identity, 0, len(rows))
	for _, r := range rows {
		items = append(items, s.rowToIdentity(r))
	}
	return items, nil
}

func (s *sqliteProvider) GetIdentity(ctx context.Context, userID, id string) (msglayer.Identity, error) {
	row, err := s.q.GetIdentity(ctx, &sqlc.GetIdentityParams{ID: id, UserID: userID})
	if err != nil {
		return msglayer.Identity{}, fmt.Errorf("identity not found: %s", id)
	}
	return s.rowToIdentity(row), nil
}

func (s *sqliteProvider) GetThread(ctx context.Context, userID, threadID string) ([]msglayer.TimelineItem, error) {
	relations, err := s.q.ListRelationsByTarget(ctx, &sqlc.ListRelationsByTargetParams{
		Type:   "same_thread",
		Target: threadID,
	})
	if err != nil {
		return nil, err
	}
	var items []msglayer.TimelineItem
	for _, rel := range relations {
		event, err := s.q.GetEvent(ctx, &sqlc.GetEventParams{ID: rel.EventID, UserID: userID})
		if err != nil {
			continue
		}
		item, err := s.eventToTimelineItem(ctx, event)
		if err != nil {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

// ==================== Users ====================

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

// ==================== Refresh Tokens ====================

func (s *sqliteProvider) SaveRefreshToken(ctx context.Context, token RefreshTokenRecord) error {
	return s.q.SaveRefreshToken(ctx, &sqlc.SaveRefreshTokenParams{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		ParentID:  sql.NullString{},
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

// ==================== Setup ====================

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

func (s *sqliteProvider) UpdateUserPasswordHash(ctx context.Context, userID, newHash string) error {
	return s.q.UpdateUserPasswordHash(ctx, &sqlc.UpdateUserPasswordHashParams{
		PasswordHash: newHash,
		PasswordSalt: "",
		ID:           userID,
	})
}

// ==================== Type conversions ====================

func (s *sqliteProvider) rowToUser(row *sqlc.User) UserRecord {
	var roles, buttons []string
	json.Unmarshal([]byte(row.Roles), &roles)
	json.Unmarshal([]byte(row.Buttons), &buttons)
	return UserRecord{
		ID:           row.ID,
		UserName:     row.UserName,
		Email:        row.Email.String,
		PasswordHash: row.PasswordHash,
		PasswordSalt: row.PasswordSalt,
		Roles:        roles,
		Buttons:      buttons,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func (s *sqliteProvider) rowToIdentity(row *sqlc.Identity) msglayer.Identity {
	var phones, emails, labels []string
	var meta map[string]any
	json.Unmarshal([]byte(row.Phones), &phones)
	json.Unmarshal([]byte(row.Emails), &emails)
	json.Unmarshal([]byte(row.Labels), &labels)
	json.Unmarshal([]byte(row.Meta), &meta)
	var avatar *string
	if row.Avatar.Valid {
		avatar = &row.Avatar.String
	}
	return msglayer.Identity{
		ID:          row.ID,
		Type:        row.Type,
		DisplayName: row.DisplayName,
		Phones:      phones,
		Emails:      emails,
		Avatar:      avatar,
		Labels:      labels,
		Meta:        meta,
	}
}

func (s *sqliteProvider) eventToTimelineItem(ctx context.Context, row *sqlc.Event) (msglayer.TimelineItem, error) {
	participants, _ := s.q.GetEventParticipants(ctx, row.ID)
	var meta map[string]any
	json.Unmarshal([]byte(row.Meta), &meta)
	return msglayer.TimelineItem{
		EventID:        row.ID,
		Type:           row.Type,
		Timestamp:      row.Timestamp.UTC().Format(time.RFC3339),
		Direction:      row.Direction,
		ContentSummary: row.ContentSummary.String,
		Participants:   participants,
		Meta:           meta,
	}, nil
}

func (s *sqliteProvider) eventsToTimelineItems(ctx context.Context, rows []*sqlc.Event) ([]msglayer.TimelineItem, error) {
	items := make([]msglayer.TimelineItem, 0, len(rows))
	for _, row := range rows {
		item, err := s.eventToTimelineItem(ctx, row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
