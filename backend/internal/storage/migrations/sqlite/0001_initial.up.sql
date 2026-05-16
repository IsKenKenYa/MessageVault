-- Commory SQLite schema v1
-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    user_name     TEXT NOT NULL UNIQUE,
    email         TEXT,
    password_hash TEXT NOT NULL,
    password_salt TEXT NOT NULL,
    roles         TEXT NOT NULL DEFAULT '["R_USER"]',
    buttons       TEXT NOT NULL DEFAULT '["view","import"]',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_user_name ON users(user_name);

-- 刷新令牌表
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    parent_id  TEXT REFERENCES refresh_tokens(id),
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens(expires_at);

-- 会话表
CREATE TABLE IF NOT EXISTS sessions (
    id               TEXT PRIMARY KEY,
    user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_id TEXT REFERENCES refresh_tokens(id) ON DELETE SET NULL,
    device_name      TEXT,
    device_type      TEXT,
    ip_address       TEXT,
    user_agent       TEXT,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at       DATETIME
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);

-- 导入记录表
CREATE TABLE IF NOT EXISTS imports (
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL,
    schema_version TEXT NOT NULL,
    is_delta       BOOLEAN NOT NULL DEFAULT FALSE,
    imported_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    source_path    TEXT,
    event_count    INTEGER NOT NULL DEFAULT 0,
    identity_count INTEGER NOT NULL DEFAULT 0,
    raw_json       TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_imports_user_id ON imports(user_id);
CREATE INDEX IF NOT EXISTS idx_imports_imported_at ON imports(imported_at);

-- 身份/联系人表
CREATE TABLE IF NOT EXISTS identities (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL,
    type         TEXT NOT NULL,
    display_name TEXT NOT NULL,
    phones       TEXT NOT NULL DEFAULT '[]',
    emails       TEXT NOT NULL DEFAULT '[]',
    avatar       TEXT,
    labels       TEXT NOT NULL DEFAULT '[]',
    meta         TEXT NOT NULL DEFAULT '{}',
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_identities_user_id ON identities(user_id);
CREATE INDEX IF NOT EXISTS idx_identities_type ON identities(type);

-- 通信事件表
CREATE TABLE IF NOT EXISTS events (
    id              TEXT PRIMARY KEY,
    user_id         TEXT NOT NULL,
    import_id       TEXT,
    type            TEXT NOT NULL,
    timestamp       DATETIME NOT NULL,
    direction       TEXT NOT NULL,
    content_summary TEXT,
    content         TEXT NOT NULL DEFAULT '{}',
    meta            TEXT NOT NULL DEFAULT '{}',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_events_user_id ON events(user_id);
CREATE INDEX IF NOT EXISTS idx_events_import_id ON events(import_id);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);
CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
CREATE INDEX IF NOT EXISTS idx_events_user_type ON events(user_id, type);
CREATE INDEX IF NOT EXISTS idx_events_user_timestamp ON events(user_id, timestamp DESC);

-- 事件参与者表
CREATE TABLE IF NOT EXISTS event_participants (
    event_id    TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    identity_id TEXT NOT NULL,
    PRIMARY KEY (event_id, identity_id)
);

CREATE INDEX IF NOT EXISTS idx_event_participants_identity ON event_participants(identity_id);

-- 事件关系表
CREATE TABLE IF NOT EXISTS relations (
    id       TEXT PRIMARY KEY,
    event_id TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    type     TEXT NOT NULL,
    target   TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_relations_event_id ON relations(event_id);
CREATE INDEX IF NOT EXISTS idx_relations_type_target ON relations(type, target);

-- 审计日志表
CREATE TABLE IF NOT EXISTS audit_log (
    id         TEXT PRIMARY KEY,
    user_id    TEXT REFERENCES users(id) ON DELETE SET NULL,
    action     TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    detail     TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log(created_at);

-- 初始化状态表
CREATE TABLE IF NOT EXISTS setup (
    id             TEXT PRIMARY KEY,
    version        TEXT NOT NULL,
    initialized_at DATETIME NOT NULL,
    usage_mode     TEXT
);

-- 迁移版本追踪
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    dirty   BOOLEAN NOT NULL DEFAULT FALSE
);
