# MsgLayer Indexing Notes

`MsgLayer` separates canonical exported data from query-time indexes.

## Canonical Source

- root export JSON
- identities
- events

## First Query Indexes

- keyword search across SMS text, voice transcript, voice summary
- participant/contact grouping
- timeline ordering by event timestamp
- thread reconstruction from `same_thread`

## Storage Direction

The first backend release supports:

- SQLite for local-first deployment
- PostgreSQL for service-style deployment

The schema is normalized enough to support future adapters without changing the exported `MsgLayer` wire format.
