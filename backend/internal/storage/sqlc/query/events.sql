-- name: CreateEvent :exec
INSERT INTO events (id, user_id, import_id, type, timestamp, direction, content_summary, content, meta)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetEvent :one
SELECT * FROM events WHERE id = ? AND user_id = ? LIMIT 1;

-- name: ListEvents :many
SELECT * FROM events
WHERE user_id = ?
  AND type = COALESCE(NULLIF(@type, ''), type)
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: SearchEvents :many
SELECT * FROM events
WHERE user_id = ?
  AND (COALESCE(content_summary, '') || ' ' || content) LIKE '%' || @search_term || '%'
  AND type = COALESCE(NULLIF(@type, ''), type)
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: TimelineEvents :many
SELECT * FROM events
WHERE user_id = ?
  AND type = COALESCE(NULLIF(@type, ''), type)
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: TimelineEventsByParticipant :many
SELECT DISTINCT e.id, e.user_id, e.import_id, e.type, e.timestamp, e.direction, e.content_summary, e.content, e.meta, e.created_at
FROM events e
JOIN event_participants ep ON e.id = ep.event_id
JOIN identities i ON ep.identity_id = i.id
WHERE e.user_id = ?
  AND e.type = COALESCE(NULLIF(@type, ''), e.type)
  AND i.display_name LIKE '%' || @participant || '%'
ORDER BY e.timestamp DESC
LIMIT ? OFFSET ?;

-- name: CountEventsByUser :one
SELECT COUNT(*) FROM events WHERE user_id = ?;

-- name: CountEventsByUserAndType :many
SELECT type, COUNT(*) AS count FROM events WHERE user_id = ? GROUP BY type;

-- name: DeleteEvent :exec
DELETE FROM events WHERE id = ? AND user_id = ?;

-- name: DeleteEventsByImport :exec
DELETE FROM events WHERE import_id = ? AND user_id = ?;

-- name: LastActivityTime :one
SELECT MAX(timestamp) FROM events WHERE user_id = ?;

-- name: InsertEventParticipant :exec
INSERT OR IGNORE INTO event_participants (event_id, identity_id) VALUES (?, ?);

-- name: GetEventParticipants :many
SELECT identity_id FROM event_participants WHERE event_id = ?;
