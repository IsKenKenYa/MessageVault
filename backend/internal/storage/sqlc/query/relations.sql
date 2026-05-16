-- name: CreateRelation :exec
INSERT INTO relations (id, event_id, type, target) VALUES (?, ?, ?, ?);

-- name: ListRelationsByEvent :many
SELECT * FROM relations WHERE event_id = ?;

-- name: ListRelationsByTarget :many
SELECT * FROM relations WHERE type = ? AND target = ?;

-- name: DeleteRelationsByEvent :exec
DELETE FROM relations WHERE event_id = ?;
