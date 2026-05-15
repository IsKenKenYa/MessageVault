# Commory Governance Notice

Commory is a local-first communication memory system. The project turns SMS, calls, contacts, and future communication sources into structured MsgLayer data that users can keep locally or optionally sync to a self-hosted server.

## Principles

- Local-first: Android must be useful without a server account or network.
- Server-optional: Commory Server adds sync, search, and remote restore; it must not become a hidden requirement for local backup.
- Privacy-controlled: logs, errors, analytics, docs, and tests must not expose message bodies, contact contents, access tokens, refresh tokens, or private backup payloads.
- Contract-driven: Android, backend, web, and MsgLayer changes that cross boundaries must update docs and tests together.

## Repository Boundaries

- `android/`: Commory Android client and SDK modules.
- `backend/`: self-hosted Commory Server.
- `web/`: Vue dashboard for server-backed workflows.
- `msglayer/`: canonical interchange schema and examples.
- `docs/`: current engineering and API documentation.
- `previewer/`: historical XML SMS viewer archive; do not update for current work.
- `references/`: read-only external reference code; do not develop features there.

## Change History

Use standard changelog and review channels:

- Durable release history goes into the relevant `CHANGELOG.md`.
- Pull requests and commits describe implementation details, testing, and AI assistance where useful.
- Do not create or update `AI_EDIT_LOG.md`; AI-assisted work is not tracked in a special side log.

## Standards

`docs/engineering-standards.md` is the canonical engineering standard. It covers naming, packages, skills, testing, logging, context management, and anti-bloat rules.

## License

The root repository is licensed under GPL v3. External code under `references/` keeps its upstream license and is not automatically part of Commory's distributable source.
