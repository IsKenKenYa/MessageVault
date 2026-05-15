# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Commory is a communication memory system that transforms SMS, call logs, and contacts into structured, queryable data. It is a monorepo with four layers:

- **android/** — Kotlin/Jetpack Compose ingestion app with multi-module SDK architecture
- **backend/** — Go HTTP server (stdlib only, zero external dependencies)
- **web/** — Vue 3 + Element Plus admin dashboard (based on art-design-pro)
- **msglayer/** — Canonical JSON Schema v0.1 interchange format (the data contract between Android and server)

The `previewer/` directory is deprecated (legacy XML SMS viewer). The `references/` directory contains read-only git submodules — never edit files there.

## Common Commands

### Backend (Go)

```bash
cd backend
go test ./... -coverprofile=coverage.out   # run all tests with coverage
go vet ./...                                 # static analysis
gofmt -l backend                             # format check (run from repo root)
```

### Android

```bash
cd android
./gradlew :app:compileDebugKotlin            # compile (CI gate)
./gradlew :app:testDebugUnitTest             # JVM unit tests (CI gate)
./gradlew :sdk:backup:test                   # pure Kotlin SDK tests (no device)
./gradlew :sdk:auth:test                     # pure Kotlin SDK tests (no device)
./gradlew :sdk:storage:connectedAndroidTest  # instrumented tests (needs device/emulator)
```

### Web Dashboard

```bash
cd web
pnpm install
pnpm dev          # dev server
pnpm build        # type-check + production build
pnpm lint         # eslint
```

### CI Governance Scripts

```bash
bash scripts/check-repo-hygiene.sh     # no tracked build artifacts, logs, IDE metadata
bash scripts/check-android-i18n.sh     # string keys complete across all locales
bash scripts/report-loc-complexity.sh  # LOC report for review attention
```

## Architecture

### Android Multi-Module SDK

```
app → sdk/backup, sdk/auth, sdk/storage
sdk/backup ✗→ sdk/auth, sdk/storage, app
sdk/auth   ✗→ sdk/backup, sdk/storage, app
sdk/storage ✗→ sdk/backup, sdk/auth, app
```

- `sdk/backup` and `sdk/auth` are **pure Kotlin/JVM** — no Android dependencies. Use `java-library` + `kotlin.jvm` plugin. All platform operations exposed via interfaces (e.g., `SmsReader`, `AuthProvider`); `app` provides Android implementations.
- `sdk/storage` is an **Android Library** — uses Room, Retrofit. Exposes pure Kotlin interfaces internally.
- `app` owns Compose UI, navigation, permissions, runtime mode setup, and all SDK interface implementations. Dependency injection via constructor — no Service Locator or global singletons.

### Runtime Modes

The Android app has two modes defined in `android/app/src/main/java/imken/messagevault/mobile/runtime/`:
- **LOCAL_ONLY** — no network, no auth, data stays on device
- **COMMORY_SERVER** — connects to backend, uses server-backed auth and storage

Local-only mode must never require backend auth. Server mode must isolate auth/session state and preserve local backup files when modes switch.

### Backend

Go server with CLI (`cmd/commory/main.go`). Key packages under `internal/`:
- `api/` — HTTP server with public/private route mux, JWT auth middleware
- `auth/` — auth service, JWT model, middleware
- `storage/` — SQLite and PostgreSQL providers, file store
- `msglayer/` — MsgLayer validator and types
- `query/` — search, timeline, identities, threads
- `importers/` — JSON importer
- `config/` — env-var-based configuration

Backend configuration is via environment variables (all prefixed `COMMORY_`). See `internal/config/config.go` for the full list.

### MsgLayer

JSON Schema v0.1 in `msglayer/schema/v0.1/`. Defines events, identities, relations, and content types (SMS, call, voice, contact snapshot). This is the canonical data contract — Android serializes to it, backend validates and imports it.

## Coding Conventions

- **Kotlin**: 4-space indent, follow [Kotlin coding conventions](https://kotlinlang.org/docs/coding-conventions.html). Files: PascalCase (classes), camelCase (utilities).
- **Vue/JS/TS**: 2-space indent, Composition API, follow [Vue style guide](https://vuejs.org/style-guide/). Files: PascalCase.
- **Go**: standard `gofmt`, no external dependencies (stdlib only).
- Line length: 120 chars recommended.
- Every file ends with a trailing newline.

## Localization (Android)

Every user-visible string must exist in three resource folders:
- `values/strings.xml` (default)
- `values-en/strings.xml` (English)
- `values-zh-rCN/strings.xml` (Simplified Chinese)

Compose UI must use `stringResource`. Dynamic errors wrapped in string resources with placeholders. CI enforces key completeness via `scripts/check-android-i18n.sh`.

## Testing Standards

| Layer | Target | Coverage |
|-------|--------|----------|
| Backend storage/auth/validators | `go test ./...` | 80%+ |
| Backend API handlers | `go test ./...` | 70%+ |
| Android SDK core logic | JVM unit tests | 80%+ |
| Android App ViewModels | JVM unit tests | 70%+ |

New SDK interfaces must include contract tests. Pure Kotlin SDK modules use JVM tests only — no Android emulator needed.

## Logging and Errors

- Logs must identify subsystem and context without exposing message bodies, contact contents, auth tokens, or refresh tokens.
- User-facing errors are typed at the ViewModel/data boundary and localized at the UI boundary.
- Server upload failures must be non-destructive when a local backup already exists.

## Versioning and Changelog

- Each sub-project maintains its own `CHANGELOG.md` following [Keep a Changelog](https://keepachangelog.com/) format.
- AI-assisted edits must be recorded in `AI_EDIT_LOG.md`.

## References Directory

`references/` contains read-only git submodules (`art-design-pro`, `new-api`) for design reference. Never modify files there. To absorb a reference implementation, rewrite it in your own module.
