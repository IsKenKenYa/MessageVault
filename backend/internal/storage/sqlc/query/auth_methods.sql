-- name: CreateAuthMethod :exec
INSERT INTO user_auth_methods (id, user_id, provider_type, provider_user_id, metadata)
VALUES (?, ?, ?, ?, ?);

-- name: GetAuthMethodByProvider :one
SELECT * FROM user_auth_methods WHERE provider_type = ? AND provider_user_id = ? LIMIT 1;

-- name: ListAuthMethodsByUser :many
SELECT * FROM user_auth_methods WHERE user_id = ?;

-- name: DeleteAuthMethod :exec
DELETE FROM user_auth_methods WHERE id = ? AND user_id = ?;
