-- name: GetSetupStatus :one
SELECT * FROM setup LIMIT 1;

-- name: SaveSetup :exec
INSERT OR REPLACE INTO setup (id, version, initialized_at, usage_mode)
VALUES (?, ?, ?, ?);

-- name: GetAuditLogs :many
SELECT * FROM audit_log
WHERE (NULLIF(@user_id, '') IS NULL OR user_id = @user_id)
  AND (NULLIF(@action, '') IS NULL OR action = @action)
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CreateAuditLog :exec
INSERT INTO audit_log (id, user_id, action, ip_address, user_agent, detail)
VALUES (?, ?, ?, ?, ?, ?);

-- name: CountAuditLogs :one
SELECT COUNT(*) FROM audit_log
WHERE (NULLIF(@user_id, '') IS NULL OR user_id = @user_id)
  AND (NULLIF(@action, '') IS NULL OR action = @action);
