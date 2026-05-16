-- Passkey 凭证表（每用户可多个）
CREATE TABLE IF NOT EXISTS passkey_credentials (
    id               TEXT PRIMARY KEY,
    user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id    TEXT NOT NULL UNIQUE,
    public_key       TEXT NOT NULL,
    attestation_type TEXT NOT NULL DEFAULT '',
    aaguid           TEXT NOT NULL DEFAULT '',
    sign_count       INTEGER NOT NULL DEFAULT 0,
    transports       TEXT NOT NULL DEFAULT '[]',
    name             TEXT NOT NULL DEFAULT '',
    last_used_at     DATETIME,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_passkey_user_id ON passkey_credentials(user_id);
CREATE INDEX IF NOT EXISTS idx_passkey_credential_id ON passkey_credentials(credential_id);

-- 认证 Challenge 临时存储（WebAuthn + OAuth state 通用）
CREATE TABLE IF NOT EXISTS auth_challenges (
    id         TEXT PRIMARY KEY,
    challenge  TEXT NOT NULL,
    user_id    TEXT,
    flow_type  TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_challenges_expires ON auth_challenges(expires_at);

-- 统一认证方式表
CREATE TABLE IF NOT EXISTS user_auth_methods (
    id               TEXT PRIMARY KEY,
    user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_type    TEXT NOT NULL,
    provider_user_id TEXT NOT NULL,
    metadata         TEXT NOT NULL DEFAULT '{}',
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_type, provider_user_id)
);

CREATE INDEX IF NOT EXISTS idx_auth_methods_user ON user_auth_methods(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_methods_provider ON user_auth_methods(provider_type, provider_user_id);
