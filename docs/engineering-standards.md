# Engineering Standards

## Architecture Boundaries

- Android app code owns Compose UI, permissions, platform readers/writers, navigation, and runtime mode setup.
- `android/sdk/backup` owns pure backup/restore orchestration and MsgLayer mapping.
- `android/sdk/auth` owns auth contracts and provider abstractions.
- `android/sdk/storage` owns local and remote storage contracts.
- Backend handlers remain thin: parse HTTP, call auth/storage/query/import services, return envelopes.
- MsgLayer schema is the canonical interchange format between Android and Commory Server.

## Localization

- Every Android user-visible string must exist in `values/strings.xml`, `values-en/strings.xml`, and `values-zh-rCN/strings.xml`.
- New Compose UI must use `stringResource`.
- Dynamic errors should be wrapped in string resources with placeholders.

## Testing

- Backend core packages target 80%+ coverage for storage, auth, validators, and import/query logic.
- Backend API handlers target 70%+ coverage for auth, setup, imports, and mobile contract behavior.
- Android domain/ViewModel behavior should be unit-tested where possible.
- Instrumented tests are reserved for Android platform integrations such as permissions, content providers, and Compose flows that require device/runtime APIs.

## Quality Gates

- Go: `gofmt`, `go vet ./...`, `go test ./...`.
- Android: `:app:compileDebugKotlin`, unit tests, lint when Android SDK tooling is available.
- Repo hygiene: no tracked `build/`, `.gradle/`, generated APK/AAB, logs, or IDE metadata.
- LOC and complexity are reported in CI for review attention. Large files or large diffs require decomposition or a short justification in the PR.

## Logging and Errors

- Logs should identify subsystem and context without exposing message bodies, contact contents, auth tokens, or refresh tokens.
- User-facing errors should be typed at the ViewModel/data boundary and localized at the UI boundary.
- Server upload failures must be non-destructive when a local backup already exists.

## Release Checklist

- Mobile API contract still matches `docs/mobile-api.md`.
- Local-only mode works without network or auth.
- Server mode validates setup, login/register, upload, list, and export flows.
- Locale keys are complete for default, English, and Simplified Chinese resources.
- Backup permissions and restore permissions are verified separately.
- Rollback notes describe any storage, session, or API compatibility change.
