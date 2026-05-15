# Engineering Standards

This is the canonical engineering standard for Commory. Keep durable rules here, keep agent workflow entrypoints in `AGENTS.md`, and keep release history in `CHANGELOG.md`.

## Architecture Boundaries

- Android `app` owns Compose UI, permissions, platform readers/writers, navigation, runtime mode setup, and Android-specific networking.
- Android `sdk/backup` owns pure Kotlin backup/restore orchestration and MsgLayer mapping.
- Android `sdk/auth` owns pure Kotlin auth contracts.
- Android `sdk/storage` owns storage contracts and Android storage implementation.
- Backend handlers stay thin: parse HTTP, call auth/storage/query/import/setup services, return envelopes.
- Web owns server-backed dashboard workflows and must follow existing Vue 3, Vite, Pinia, Element Plus patterns.
- MsgLayer schema is the canonical interchange format between Android, backend, web, and future agents.

## Naming

- Android application id: `com.iskenkenya.commory`.
- Android app namespace and source packages: `com.iskenkenya.commory.mobile`.
- Android SDK packages: `com.iskenkenya.commory.sdk.*`.
- Backend Go module remains `github.com/IsKenKenYa/Commory/backend`.
- Current user-visible product name is `Commory`; old names may remain only in historical changelog entries.

## Repository Boundaries

- `previewer/` is a historical archive. Do not update it for current roadmap, CI, docs, or naming work.
- `references/` is read-only external reference code.
- `.agents/skills` is the source of truth for project skills.
- `.claude/skills` is generated from `.agents/skills` for Claude Code compatibility.
- Do not add tracked build output, logs, `.DS_Store`, IDE files, debug reports, or `AI_EDIT_LOG.md`.

## Skills

Skills are context-loading tools, not general documentation folders.

- Required per skill: `SKILL.md` with concise frontmatter (`name`, `description`) and a lean workflow.
- Recommended: `agents/openai.yaml` when UI metadata is needed.
- Optional: `scripts/`, `references/`, `assets/`.
- Forbidden inside skills: `README.md`, `CHANGELOG.md`, `AGENTS.md`, install guides, quick references, and duplicated long-form docs.

When updating skills, edit `.agents/skills`, then run:

```bash
bash scripts/sync-agent-skills.sh
```

Use `bash scripts/sync-agent-skills.sh --check` in CI or review to ensure `.claude/skills` has not drifted.

## Context Management

Use the smallest context bundle needed:

- Android UI: navigation host, target screen, string resources, related ViewModel.
- Android auth/network: runtime environment, auth provider, server client, SDK contracts, mobile API docs.
- Backend API: server handler, auth middleware/service, storage provider, API tests, mobile API docs.
- Web: API client, auth store, route guards, target view, existing dashboard conventions.
- MsgLayer: schema, examples, validators, Android mapper/serializer.
- Governance: this file, `AGENTS.md`, CI workflow, scripts, and skills.

Avoid new shared abstractions unless they remove real duplication or clarify a cross-module contract. Large files or large diffs require decomposition or a short PR justification.

## Localization

- Every Android user-visible string must exist in `values/strings.xml`, `values-en/strings.xml`, and `values-zh-rCN/strings.xml`.
- Compose UI must use `stringResource`.
- Dynamic errors should be typed before the UI boundary and localized with placeholders.

## Logging And Privacy

- Logs may identify subsystem and operation context.
- Logs must not include message bodies, contact contents, auth tokens, refresh tokens, backup payloads, or private server URLs with credentials.
- Server upload failures must be non-destructive when a local backup already exists.
- Local-only mode must not require backend auth or network.

## Testing

- Backend core packages target strong coverage for storage, auth, validators, import/query logic, setup, and mobile contracts.
- Android SDK pure Kotlin logic should use JVM tests.
- Android ViewModel/domain behavior should use JVM tests where possible.
- Instrumented tests are reserved for permissions, content providers, Compose flows, and platform behavior.
- Web changes must pass lint and production build.

## Quality Gates

```bash
cd backend && go vet ./... && go test ./... -coverprofile=coverage.out
cd android && ./gradlew :app:compileDebugKotlin && ./gradlew :app:testDebugUnitTest
cd web && pnpm install --frozen-lockfile && pnpm lint && pnpm build
bash scripts/check-repo-hygiene.sh
bash scripts/check-android-i18n.sh
bash scripts/sync-agent-skills.sh --check
```

## Release Checklist

- Mobile API contract still matches `docs/mobile-api.md`.
- Local-only mode works without network or auth.
- Server mode validates setup, login/register, refresh, logout, upload, list, and export flows.
- Android locale keys are complete.
- Cross-boundary API/schema changes include docs and tests.
- Changelog entries describe user-facing and compatibility changes.
