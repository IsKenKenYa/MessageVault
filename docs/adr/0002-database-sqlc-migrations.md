# ADR 0002: SQLC and Versioned Migrations for Storage

## Status

Accepted.

## Context

Commory currently exposes `sqlite` and `postgres` providers, but both are backed by the same JSON file store adapter. That was useful for early product flow, but it is not a real multi-database foundation.

The long-term target is SQLite for personal/local deployments and Postgres or MySQL for production or multi-user deployments.

## Decision

Commory will use explicit SQL plus generated types for the durable storage layer. The preferred route is `sqlc` with versioned migrations for SQLite, Postgres, and MySQL.

Cross-database differences must be visible in query and migration organization instead of being hidden in scattered runtime branches. When behavior differs by database, the storage contract test should define the shared behavior and each driver should prove it.

The existing JSON file store remains useful as a development fixture and migration source, but it must not be presented as a real SQLite or Postgres implementation.

## Consequences

- Add storage contract tests before replacing the current file-backed provider.
- Introduce schema migrations per supported database family.
- Generate typed query packages from SQL instead of introducing a business ORM.
- Keep setup and diagnostics explicit about the active database driver and persistence risk.

