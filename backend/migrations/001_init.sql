CREATE TABLE imports (
  id TEXT PRIMARY KEY,
  schema_version TEXT NOT NULL,
  imported_at TEXT NOT NULL,
  source_path TEXT NOT NULL,
  raw_json TEXT NOT NULL
);
