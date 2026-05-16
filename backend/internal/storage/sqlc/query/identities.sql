-- name: CreateIdentity :exec
INSERT INTO identities (id, user_id, type, display_name, phones, emails, avatar, labels, meta)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetIdentity :one
SELECT * FROM identities WHERE id = ? AND user_id = ? LIMIT 1;

-- name: ListIdentities :many
SELECT * FROM identities
WHERE user_id = ?
  AND type = COALESCE(NULLIF(@type, ''), type)
ORDER BY display_name ASC;

-- name: UpdateIdentity :exec
UPDATE identities
SET display_name = ?, phones = ?, emails = ?, avatar = ?, labels = ?, meta = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND user_id = ?;

-- name: DeleteIdentity :exec
DELETE FROM identities WHERE id = ? AND user_id = ?;

-- name: CountIdentitiesByUser :one
SELECT COUNT(*) FROM identities WHERE user_id = ?;

-- name: SearchIdentities :many
SELECT * FROM identities
WHERE user_id = ?
  AND (display_name || ' ' || phones) LIKE '%' || @query || '%'
ORDER BY display_name ASC;
