# Commory Reference Catalog

This directory is for architecture study only. Reference projects are not implementation dependencies, and their code, text, schemas, field names, or proprietary structures must not be copied into Commory.

## Policy

- Prefer permissive references: MIT, Apache-2.0, BSD.
- Treat AGPL, GPL, LGPL, proprietary, and unknown-license projects as read-only architecture material.
- Record why every reference exists before using it in design work.
- Convert lessons into Commory-owned ADRs, interfaces, and tests before writing product code.
- Do not import from `references/` in production or test code.

## Catalog

| Path | Source | License policy | Learn from | Do not borrow |
| --- | --- | --- | --- | --- |
| `references/ClaudeCodeSource` | Local ignored source snapshot | Restricted / clean-room only | Agent product architecture, task/session UX, tool permissions, command surfaces, compact/context concepts, bridge and resume patterns | Source code, strings, field names, schemas, proprietary structures |
| `references/GitHubDaily` | `https://github.com/GitHubDaily/GitHubDaily.git` | Unknown / discovery index only | Finding Go/self-hosted/database/auth/agent/concurrency projects for later review | Code or article text without separate license validation |
| `references/new-api` | `https://github.com/QuantumNous/new-api.git` | AGPL read-only architecture reference | Multi-database awareness, setup UX, user/admin/root auth layering, operational warnings | Implementation, handlers, model names, field names, SQL, UI text |
| `references/hermes-agent` | `https://github.com/NousResearch/hermes-agent.git` | MIT permissive reference | Provider registry, plugin lifecycle, config validation, diagnostics, long-running agent workflows | Direct Python implementation or project-specific prompts |
| `references/openclaw` | `https://github.com/openclaw/openclaw.git` | MIT permissive reference | Plugin-agnostic core boundaries, extension governance, protocol/versioning habits | Product-specific plugin contracts or wording |
| `references/art-design-pro` | `https://github.com/Daymychen/art-design-pro.git` | MIT permissive reference | Admin UI layout, theme organization, dashboard ergonomics | Branding and app-specific UI copy |
| `references/memos` | `https://github.com/usememos/memos.git` | MIT permissive reference | Go self-hosted product structure, SQLite-first operation, lightweight domain modeling | Exact data models, API shapes, UI strings |
| `references/gitea` | `https://github.com/go-gitea/gitea.git` | MIT permissive reference | Mature Go self-hosted governance, multi-database operations, permissions, upgrade discipline | Git service domain implementation, templates, naming |
| `references/pocketbase` | `https://github.com/pocketbase/pocketbase.git` | MIT permissive reference | Single-binary SQLite-first architecture, auth, hooks, realtime ergonomics | Framework internals or public API names as Commory APIs |
| `references/sqlc` | `https://github.com/sqlc-dev/sqlc.git` | MIT permissive reference | Type-safe SQL generation, query organization, migration-adjacent workflow | Tool internals |

## GitHubDaily Mining

Use `scripts/mine-githubdaily-references.sh` to extract candidate repositories from the local GitHubDaily index. A candidate can become a submodule only after its language, license, and Commory learning value are recorded here.

