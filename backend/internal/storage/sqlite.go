package storage

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	sqlc "github.com/IsKenKenYa/Commory/backend/internal/storage/sqlc/gen"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/sqlite/0001_initial.up.sql
var sqliteMigration001 string

//go:embed migrations/sqlite/0002_auth_hardening.up.sql
var sqliteMigration002 string

var sqliteMigrations = []struct {
	Version int
	SQL     string
}{
	{1, sqliteMigration001},
	{2, sqliteMigration002},
}

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
	// 确保 schema_migrations 表存在
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version BIGINT PRIMARY KEY, dirty BOOLEAN NOT NULL DEFAULT FALSE)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	// 获取当前已应用的最高版本
	currentVersion := 0
	row := s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`)
	_ = row.Scan(&currentVersion)

	// 按顺序应用未执行的迁移
	for _, m := range sqliteMigrations {
		if m.Version <= currentVersion {
			continue
		}
		if _, err := s.db.ExecContext(ctx, m.SQL); err != nil {
			return fmt.Errorf("run migration %04d: %w", m.Version, err)
		}
		if _, err := s.db.ExecContext(ctx, `INSERT OR REPLACE INTO schema_migrations (version, dirty) VALUES (?, FALSE)`, m.Version); err != nil {
			return fmt.Errorf("record migration %04d: %w", m.Version, err)
		}
	}

	return nil
}
