# Commory Architecture Roadmap

Commory is moving from prototype storage and direct handlers toward durable, testable service boundaries. The near-term work is intentionally boring: database correctness, authentication hardening, and architecture guardrails before richer agent behavior.

## Phase 1: Durable Foundation

- Replace the file-backed `sqlite` and `postgres` provider aliases with real SQL implementations.
- Use `sqlc` plus versioned migrations for SQLite, Postgres, and MySQL.
- Add storage contract tests for auth, setup, import/export, search, timeline, identities, and refresh token rotation.
- Make setup diagnostics explicit about database type and SQLite persistence risks.

## Phase 2: Authentication and Permissions

- Keep short-lived access tokens with refresh token rotation.
- Add session/device records, revoke audit, and typed role constants for user/admin/root behavior.
- Keep HTTP middleware thin; auth decisions belong in service-level code.

## Phase 3: Agent Runtime

- Build the Go-owned runtime boundary in `backend/internal/agent`.
- Add provider adapters behind interfaces.
- Add a tool registry and permission broker before allowing tools to mutate local state.
- Persist task/session state through storage interfaces.

## Phase 4: Reference Mining

- Use GitHubDaily to discover candidate Go projects.
- Prefer MIT, Apache-2.0, and BSD references.
- Use AGPL/GPL/LGPL/unknown-license projects only for read-only architecture observation.
- Record every accepted reference in `references/README.md`.

