# Commory

> Turn communication into memory.

Commory is a local-first communication memory system. It turns SMS, call logs, contacts, and future communication sources into structured, queryable, syncable, AI-ready personal data assets.

Commory is not a backup-file-only tool. Backup is the entry point; the goal is a durable data layer for personal communication history.

## Architecture

```text
Commory
├── android/     # Android client: local backup, restore, optional server sync
├── backend/     # Commory Server: auth, imports, query, self-hosted API
├── web/         # Vue dashboard for server-backed workflows
├── msglayer/    # Canonical cross-platform communication schema
├── docs/        # Current engineering and API documentation
├── scripts/     # CI and governance scripts
├── previewer/   # Historical XML SMS viewer archive; do not update
└── references/  # Read-only external reference code
```

The current product state is local backup plus optional server upload. The next step is bidirectional sync: upload, remote restore, and incremental sync. Long-term directions include end-to-end encryption, multi-device sync, and AI Agent context.

The core principles remain: local-first, server-optional, privacy-controlled.

## Components

- `android/`: Commory Android client. App id is `com.iskenkenya.commory`; app source namespace is `com.iskenkenya.commory.mobile`; SDK namespaces are `com.iskenkenya.commory.sdk.*`.
- `backend/`: self-hosted Commory Server in Go. The mobile API contract lives in `docs/mobile-api.md`.
- `web/`: Vue 3, Vite, and Element Plus dashboard.
- `msglayer/`: canonical JSON Schema interchange format.
- `previewer/`: historical XML SMS viewer archive. It is kept as old code and is not part of current CI, docs, or roadmap work.

## Quick Start

```bash
cd android
./gradlew :app:compileDebugKotlin
./gradlew :app:testDebugUnitTest
```

```bash
cd backend
go test ./...
go run ./cmd/commory
```

```bash
cd web
pnpm install
pnpm dev
```

## Governance

- Agent entrypoint: `AGENTS.md`
- Engineering standards: `docs/engineering-standards.md`
- Governance notice: `NOTICE.md`
- Mobile API contract: `docs/mobile-api.md`

`.agents/skills` is the source of truth for project skills. `.claude/skills` is the Claude Code compatibility mirror. After changing skills, run:

```bash
bash scripts/sync-agent-skills.sh
```

## License

The root repository is licensed under [GNU General Public License v3.0](LICENSE). External references under `references/` keep their upstream licenses.
