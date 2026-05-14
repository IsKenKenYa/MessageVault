# Android -> MsgLayer v0.1 Mapping

This document defines the first stable mapping from the current Android backup models to `msglayer/v0.1`.

## Root

- `version` -> fixed `msglayer/v0.1`
- `exported_at` -> export wall-clock time in RFC3339 UTC
- `source.platform` -> `android`
- `source.device_id` -> Android device identifier passed in by the app layer
- `source.app_version` -> Android app version passed in by the app layer

## Self Identity

Every export creates a stable self identity:

- `id` -> `self/<device_id>`
- `type` -> `device`
- `display_name` -> device info string when available
- `labels` -> `["self"]`

## Contact -> Identity

Android `Contact` maps to a MsgLayer `identity` with:

- `type` -> `person`
- `phones` -> `phoneNumbers`
- `emails` -> `emails`
- `avatar` -> `photoData`
- `meta.source` -> `contacts`

Stable identity IDs are generated from normalized phone numbers when possible, and fall back to contact ID/name fingerprints when no phone number exists.

## Contact -> contact_snapshot Event

Every exported identity from contacts also generates a `contact_snapshot` event:

- `direction` -> `system`
- `participants` -> the contact identity itself
- `content.identity_id` -> the identity ID
- `content.snapshot` -> the identity payload
- `relations` -> `references_identity`

## Message -> sms Event

- `Message.id` -> event ID suffix
- `Message.date` -> event timestamp
- `Message.type` -> `direction`
  - `1` => `inbound`
  - `2` => `outbound`
  - `3/4/5/6` => exported as `outbound` for schema compatibility while preserving the original Android type in `meta.raw_type`
  - fallback => `inbound`
- `Message.address` -> participant identity resolution input
- `Message.body` -> `content.text`
- `Message.threadId` -> `relations.same_thread`
- `Message.readState` -> `meta.read`
- `Message.messageStatus` -> `meta.status`

## CallLog -> call Event

- `CallLog.id` -> event ID suffix
- `CallLog.date` -> event timestamp
- `CallLog.number` / `CallLog.contact` -> participant identity resolution input
- `CallLog.type` -> `direction` and `content.call_type`
- `CallLog.duration` -> `content.duration_sec`
- `CallLog.contact` -> `meta.contact_name`

Timestamp ordering in exports relies on the mapper normalizing all event timestamps to UTC RFC3339 before sorting.

## Restore Boundary

`MsgLayer` is now the primary export format.

Current restore compatibility is maintained by converting `MsgLayer` events back into the existing Android restore DTOs:

- `sms` -> `SmsData`
- `call` -> `CallLogData`
- `identity/contact_snapshot` -> `ContactData`

This keeps restore behavior working while the write path moves to the new standard.

## Known Data Loss On Round-Trip

The `MsgLayer -> legacy restore DTO` bridge intentionally preserves compatibility over perfect fidelity in v0.1.

Known losses today:

- contact `groups`, `websites`, and `note` remain in identity metadata only
- avatar/photo data is not restored into the legacy contact DTO
- legacy Android raw message/call types remain metadata, not first-class restore fields
- non-numeric event IDs fall back to derived numeric IDs during restore bridging
