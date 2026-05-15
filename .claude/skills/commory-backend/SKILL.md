---
name: commory-backend
description: Go backend for Commory — JWT auth, file-backed storage, MsgLayer import/query, Setup wizard, and mobile API contract.
---

# Commory Backend

Work on the Commory Go backend server.

## Architecture

```
cmd/commory/main.go          — CLI entry (serve / import / validate)
internal/cli/root.go          — Cobra commands, boot sequence
internal/api/server.go        — HTTP handlers, public/private mux
internal/auth/service.go      — PBKDF2-HMAC-SHA256 passwords, JWT HS256, login rate-limit, refresh rotation
internal/auth/middleware.go   — Bearer token extraction, context propagation
internal/setup/service.go     — First-run setup wizard, CheckSetup migration
internal/storage/storage.go   — Provider interface (GetSetupStatus, SaveSetup, HasAdminUser, UpdateUserPasswordHash, etc.)
internal/storage/filestore.go — JSON-file-backed Provider, scopeKey isolation, persist-on-write
internal/storage/sqlite.go    — SQLite adapter factory
internal/msglayer/validator.go— Schema validation + additionalProperties check
internal/msglayer/types.go    — Go structs for MsgLayer v0.1
internal/importers/           — JSON import pipeline
internal/query/               — Timeline, search, identities, threads
internal/config/config.go     — Env-driven config (AUTH_SECRET, TLS, ENV, DB_DSN, etc.)
```

## Key Patterns

- Handlers are thin: parse request, call service, return envelope `{code, msg, data}`.
- Auth middleware protects `/api/` routes; `/api/auth/` and `/api/setup` are public.
- Password hashes use `pbkdf2$` prefix; legacy `sha256$` auto-migrates on login.
- Refresh tokens are one-use (ConsumeRefreshToken revokes on refresh).
- Logout revokes the refresh token before clearing the cookie.
- Setup wizard: GET returns status, POST creates admin + writes SetupRecord. Repeat POST returns 400.
- scopeKey uses SHA-1 first 8 bytes with legacy 4-byte fallback lookup.
- Cookie Secure flag reads COMMORY_TLS; COMMORY_ENV=production rejects default auth secret.

## Workflow

1. Read `backend/internal/api/server.go` and the relevant service file.
2. Read `docs/mobile-api.md` for contract alignment.
3. Implement changes following the thin-handler pattern.
4. Run: `cd backend && go vet ./... && go test ./... -coverprofile=coverage.out`

## References

- See `references/api-routes.md` for the full route table.
- See `docs/mobile-api.md` for the mobile API contract.
