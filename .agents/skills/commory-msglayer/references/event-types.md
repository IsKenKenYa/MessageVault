# MsgLayer Event Types

## sms

Content keys: `text`

Direction: inbound, outbound

## call

Content keys: `duration_sec`, `call_type`

Direction: inbound, outbound, missed

call_type values: incoming, outgoing, missed, rejected

## voice

Content keys: `file`, `transcript`, `summary`

Direction: inbound, outbound

## contact_snapshot

Content keys: `identity_id`

Direction: none (not applicable)

## Relation Types

| Type | Status | Description |
|------|--------|-------------|
| same_thread | active | Links events in the same conversation thread |
| references_identity | active | Links an event to a contact identity |
| reply_to | reserved | Indicates a reply relationship between events |
| derived_from | reserved | Indicates one event was derived from another |

## Android Direction Mapping

| Android SMS Type | MsgLayer Direction |
|-----------------|-------------------|
| 1 (inbox) | inbound |
| 2 (sent) | outbound |
| 3 (draft) | outbound (best approximation) |
| 4 (outbox) | outbound |
| 5 (failed) | outbound |
| 6 (queued) | outbound |

| Android Call Type | MsgLayer Direction | MsgLayer call_type |
|------------------|-------------------|-------------------|
| 1 (incoming) | inbound | incoming |
| 2 (outgoing) | outbound | outgoing |
| 5 (rejected) | missed | rejected |
