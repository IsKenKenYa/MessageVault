---
name: commory-msglayer
description: MsgLayer schema v0.1 — the canonical interchange format between Android, Go backend, web, and future agents.
---

# MsgLayer Schema

Work on the MsgLayer interchange format and its implementations.

## Schema Location

```
msglayer/schema/v0.1/root.schema.json   — JSON Schema root
msglayer/examples/                       — Example export files
backend/internal/msglayer/               — Go validator, types, constants
android/sdk/backup/                      — Kotlin mapper, reader, writer
```

## Core Types

- `RootExport`: version, source, identities[], events[], relations[], indexes[]
- `Identity`: id, type, display_name, avatar?, phones[], emails[], labels[]
- `Event`: id, timestamp, type, direction, participants[], content{}, relations[], meta?
- `Relation`: type (same_thread, references_identity, reply_to, derived_from), target
- `SearchParams`: userID, keyword, contactID, type, participant, from, to, limit, offset

## Version

- Current: `msglayer/v0.1`
- Go constant: `msglayer.Version`
- Kotlin constant: `MSG_LAYER_VERSION`
- Both must stay in sync; CI does not currently check this.

## Validation Rules

- `additionalProperties: false` is enforced by the Go validator for events, identities, and content maps.
- Allowed event fields: id, timestamp, type, direction, participants, content, relations, meta.
- Allowed identity fields: id, type, display_name, avatar, phones, emails, labels, meta.
- Content keys are type-specific: sms→text, call→duration_sec/call_type, voice→file/transcript/summary, contact_snapshot→identity_id.
- Event types: sms, call, voice, contact_snapshot.
- Direction values: inbound, outbound, missed (for calls).
- Relation types: same_thread, references_identity (active); reply_to, derived_from (reserved).

## Workflow

1. Read `msglayer/schema/v0.1/root.schema.json` for the authoritative schema.
2. Read `msglayer/examples/` for valid export files.
3. Modify Go validator in `backend/internal/msglayer/validator.go` if adding new validation rules.
4. Modify Kotlin mapper in `android/sdk/backup/` if adding new event types or fields.
5. Run backend tests: `cd backend && go test ./internal/msglayer/ -v`
6. Run Android tests: `cd android && ./gradlew :sdk:backup:testDebugUnitTest`

## References

- See `references/event-types.md` for event type definitions and content schemas.
