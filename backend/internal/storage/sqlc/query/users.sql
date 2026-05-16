-- name: CreateUser :exec
INSERT INTO users (id, user_name, email, password_hash, password_salt, roles, buttons)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: FindUserByUserName :one
SELECT * FROM users WHERE user_name = ? LIMIT 1;

-- name: GetUser :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: UpdateUserPasswordHash :exec
UPDATE users SET password_hash = ?, password_salt = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: HasAdminUser :one
SELECT COUNT(*) > 0 AS has_admin FROM users WHERE roles LIKE '%R_ADMIN%' OR roles LIKE '%R_SUPER%';

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC;

-- name: UpdateUserRoles :exec
UPDATE users SET roles = ?, buttons = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;
