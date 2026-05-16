-- name: CreateImport :exec
INSERT INTO imports (id, user_id, schema_version, is_delta, source_path, event_count, identity_count, raw_json)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListImportsByUser :many
SELECT id, user_id, schema_version, is_delta, imported_at, source_path, event_count, identity_count
FROM imports
WHERE user_id = ?
ORDER BY imported_at DESC;

-- name: GetImport :one
SELECT * FROM imports WHERE id = ? AND user_id = ? LIMIT 1;

-- name: ExportImport :one
SELECT raw_json FROM imports WHERE id = ? AND user_id = ? LIMIT 1;

-- name: LatestImportID :one
SELECT id FROM imports WHERE user_id = ? ORDER BY imported_at DESC LIMIT 1;

-- name: DeleteImport :exec
DELETE FROM imports WHERE id = ? AND user_id = ?;

-- name: CountImportsByUser :one
SELECT COUNT(*) FROM imports WHERE user_id = ?;
