-- name: SaveRefreshToken :exec
INSERT INTO refresh_tokens (id, user_id, token_hash, parent_id, expires_at)
VALUES (?, ?, ?, ?, ?);

-- name: FindRefreshTokenByHash :one
SELECT * FROM refresh_tokens
WHERE token_hash = ? AND revoked_at IS NULL AND expires_at > CURRENT_TIMESTAMP
LIMIT 1;

-- name: ConsumeRefreshToken :one
UPDATE refresh_tokens
SET revoked_at = CURRENT_TIMESTAMP
WHERE token_hash = ? AND revoked_at IS NULL AND expires_at > CURRENT_TIMESTAMP
RETURNING *;

-- name: RevokeRefreshTokensByUser :exec
UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE user_id = ? AND revoked_at IS NULL;

-- name: RevokeRefreshTokenFamily :exec
WITH RECURSIVE family(id) AS (
    SELECT id FROM refresh_tokens WHERE id = ?
    UNION
    SELECT rt.id
    FROM refresh_tokens rt
    JOIN family f ON rt.parent_id = f.id
)
UPDATE refresh_tokens
SET revoked_at = CURRENT_TIMESTAMP
WHERE id IN (SELECT id FROM family);

-- name: RevokeRefreshTokenByID :exec
UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE id = ? AND revoked_at IS NULL;

-- name: CleanupExpiredTokens :exec
DELETE FROM refresh_tokens WHERE expires_at < CURRENT_TIMESTAMP;

-- name: FindAnyRefreshTokenByHash :one
SELECT * FROM refresh_tokens
WHERE token_hash = ?
LIMIT 1;
