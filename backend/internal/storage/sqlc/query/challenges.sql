-- name: CreateChallenge :exec
INSERT INTO auth_challenges (id, challenge, user_id, flow_type, expires_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetChallenge :one
SELECT * FROM auth_challenges WHERE id = ? AND expires_at > CURRENT_TIMESTAMP LIMIT 1;

-- name: DeleteChallenge :exec
DELETE FROM auth_challenges WHERE id = ?;

-- name: CleanupExpiredChallenges :exec
DELETE FROM auth_challenges WHERE expires_at < CURRENT_TIMESTAMP;
