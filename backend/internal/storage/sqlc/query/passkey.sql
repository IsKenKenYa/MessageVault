-- name: CreatePasskeyCredential :exec
INSERT INTO passkey_credentials (id, user_id, credential_id, public_key, attestation_type, aaguid, sign_count, transports, name)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetPasskeyByCredentialID :one
SELECT * FROM passkey_credentials WHERE credential_id = ? LIMIT 1;

-- name: ListPasskeysByUser :many
SELECT * FROM passkey_credentials WHERE user_id = ? ORDER BY created_at DESC;

-- name: UpdatePasskeyLastUsed :exec
UPDATE passkey_credentials SET last_used_at = CURRENT_TIMESTAMP, sign_count = ? WHERE id = ?;

-- name: DeletePasskey :exec
DELETE FROM passkey_credentials WHERE id = ? AND user_id = ?;
