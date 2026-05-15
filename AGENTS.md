# Agent Operating Spec

Commory is a local-first communication memory monorepo. Keep context narrow, respect user changes, and verify the exact surface you touched.

## Read First

- Android UI or mobile product work: `android/README.md`, `android/app/src/main/java/com/iskenkenya/commory/mobile/ui/navigation/NavigationHost.kt`, `android/app/src/main/java/com/iskenkenya/commory/mobile/runtime/`, and the target screen or ViewModel.
- Android data/auth/storage work: `android/sdk/backup`, `android/sdk/auth`, `android/sdk/storage`, plus `android/app/src/main/java/com/iskenkenya/commory/mobile/remote/`.
- Backend API work: `backend/internal/api/server.go`, `backend/internal/auth`, `backend/internal/storage`, and `docs/mobile-api.md`.
- Web dashboard work: `web/package.json`, `web/src/api`, `web/src/router`, and the target view/store module.
- MsgLayer/schema work: `msglayer/schema/v0.1/root.schema.json`, `msglayer/examples`, and `backend/internal/msglayer`.
- Governance or CI work: `docs/engineering-standards.md`, `.github/workflows/ci.yml`, `scripts/`, `.agents/skills`, and `.claude/skills`.

## Context Bundles

- Android UI bundle: navigation host, target screen, resources in `values/`, `values-en/`, `values-zh-rCN/`, and related ViewModel.
- Android auth/network bundle: runtime environment, auth provider, server client, SDK auth/storage contracts, and mobile API docs.
- Backend API bundle: server handler, auth middleware/service, storage provider, API tests, and mobile API contract.
- Web dashboard bundle: API client, auth store, route guard, target view, and Element Plus patterns already used in `web/`.
- MsgLayer bundle: schema, examples, validators, Android mapper/serializer.
- Governance bundle: engineering standards, CI workflow, repo hygiene, skills sync, i18n check.

Load only the bundle needed for the task. Do not spread temporary decisions into tool-specific files; durable rules belong in `docs/engineering-standards.md`, agent entrypoints in this file, and release history in `CHANGELOG.md`.

## Repository Boundaries

- `previewer/` is a historical archive. Do not update it for current Commory work.
- `references/` is read-only external reference code. Rewrite ideas in Commory modules instead of editing reference mirrors.
- `.agents/skills` is the source of truth for project skills.
- `.claude/skills` is a generated compatibility mirror for Claude Code. Do not edit it by hand; run `bash scripts/sync-agent-skills.sh`.
- Do not add `AI_EDIT_LOG.md`, debug reports, tracked logs, build outputs, `.DS_Store`, or IDE/cache files.

## Naming

- Android app id: `com.iskenkenya.commory`.
- Android app namespace and source packages: `com.iskenkenya.commory.mobile`.
- Android SDK packages: `com.iskenkenya.commory.sdk.*`.
- Go backend module remains `github.com/IsKenKenYa/Commory/backend`; it does not conflict with Android package names.
- User-visible current product name is `Commory`. Historical changelog entries may mention old names as history.

## Skills

Project skills must follow the Codex skill shape:

- Required: `SKILL.md` with concise YAML frontmatter (`name`, `description`) and a lean workflow.
- Recommended: `agents/openai.yaml` when UI metadata is useful.
- Optional: `scripts/`, `references/`, `assets/`.
- Forbidden inside a skill: `README.md`, `CHANGELOG.md`, `AGENTS.md`, install guides, quick references, or long duplicated docs.

When creating or updating a skill, keep `SKILL.md` short, move detailed variants to directly linked files under `references/`, and run `bash scripts/sync-agent-skills.sh`.

## Verification

- Backend: `cd backend && go vet ./... && go test ./... -coverprofile=coverage.out`.
- Android: `cd android && ./gradlew :app:compileDebugKotlin && ./gradlew :app:testDebugUnitTest`.
- Web: `cd web && pnpm install --frozen-lockfile && pnpm lint && pnpm build`.
- Governance: `bash scripts/check-repo-hygiene.sh`, `bash scripts/check-android-i18n.sh`, `bash scripts/sync-agent-skills.sh --check`, and `bash scripts/report-loc-complexity.sh`.
