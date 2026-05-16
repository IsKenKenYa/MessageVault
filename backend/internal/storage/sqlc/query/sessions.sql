-- name: CreateSession :exec
INSERT INTO sessions (id, user_id, refresh_token_id, device_name, device_type, ip_address, user_agent)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetSession :one
SELECT * FROM sessions WHERE id = ? AND revoked_at IS NULL LIMIT 1;

-- name: GetSessionByRefreshTokenID :one
SELECT * FROM sessions WHERE refresh_token_id = ? AND revoked_at IS NULL LIMIT 1;

-- name: ListSessionsByUser :many
SELECT * FROM sessions WHERE user_id = ? AND revoked_at IS NULL ORDER BY last_seen_at DESC;

-- name: RevokeSession :exec
UPDATE sessions SET revoked_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: RevokeOtherSessions :exec
UPDATE sessions SET revoked_at = CURRENT_TIMESTAMP WHERE user_id = ? AND id != ? AND revoked_at IS NULL;

-- name: UpdateSessionLastSeen :exec
UPDATE sessions SET last_seen_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: UpdateSessionRefreshToken :exec
UPDATE sessions
SET refresh_token_id = ?, last_seen_at = CURRENT_TIMESTAMP
WHERE id = ? AND revoked_at IS NULL;
