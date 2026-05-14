# MsgLayer Versioning

## Current Version

- `msglayer/v0.1`

## Rules

- Backward-compatible additive fields stay within the same major/minor version line.
- Breaking structural changes require a new schema version.
- Schema files, examples, Android mappers, and Go types must evolve together.
- `indexes` is optional derived data and must never become the canonical source of truth.

## Compatibility Intent

`v0.1` is the first stable interchange format for Commory.

- Android export switches to `MsgLayer`
- Go backend import/validation targets `MsgLayer`
- legacy backup JSON remains restore-compatible only through bridge converters

## Relation Type Status In v0.1

- Active in current code paths: `same_thread`, `references_identity`
- Reserved for future producers/consumers: `reply_to`, `derived_from`
