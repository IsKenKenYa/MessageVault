---
name: commory-android
description: Android app for Commory — dual runtime modes, Compose UI, MsgLayer backup/restore, server sync, and agent privacy model.
---

# Commory Android

Work on the Commory Android app (package `com.iskenkenya.commory.mobile`).

## Architecture

```
app/src/main/java/com/iskenkenya/commory/mobile/
  runtime/                     — AppEnvironment, RuntimeMode, RuntimeModePolicy, AppEnvironmentManager, LocaleResolver
  remote/                      — CommoryServerClient, CommoryApiService, NetworkError, DTOs
  auth/                        — CommoryServerAuthProvider
  agent/                       — AgentContextModels, AgentPrivacyPolicy
  ui/screens/                  — ModeSelectionScreen, ServerSetupScreen, BackupScreen, RestoreScreen, MoreScreen
  ui/navigation/NavigationHost — Two-level route guard (mode → auth → main)
  ui/viewmodels/               — AppViewModel, BackupViewModel, RestoreViewModel
  di/AppContainer              — Manual DI

sdk/backup/                    — Pure Kotlin MsgLayer mapper, reader, writer
sdk/auth/                      — AuthProvider interface, AuthCredentials sealed class
sdk/storage/                   — Storage contracts
```

## Key Patterns

- `AppEnvironment` is the single source of truth; persisted via DataStore; exposed as `StateFlow`.
- `currentSnapshot()` reads a `@Volatile` cache updated by `onEach`; no `runBlocking`.
- `RuntimeModePolicy` centralizes mode-dependent decisions (auth required, upload allowed, etc.).
- `CommoryServerClient` caches Retrofit services per base URL; uses separate `refreshClient` for auth endpoints.
- OkHttp `authenticator` auto-refreshes on 401; `/api/auth/` paths are excluded to prevent loops.
- `NetworkError` sealed class maps IOException subtypes to typed errors for UI localization.
- `AuthSession.isAuthenticated` checks JWT `exp` claim; `accessTokenExpiresAt()` decodes Base64 payload.
- Backup always writes local first; upload is gated by `RuntimeModePolicy.canUploadBackup()`.
- Mode switch shows `AlertDialog` confirmation; switching clears session but preserves local files.
- `AgentPrivacyPolicy` maps `RuntimeMode` to `AgentProviderPolicy` (LOCAL_ONLY vs SERVER_ALLOWED).
- Default server URL detects emulator via `Build.FINGERPRINT`.

## Workflow

1. Read `NavigationHost.kt` and the target screen/ViewModel.
2. Read `docs/android-runtime-modes.md` for mode behavior.
3. Read `docs/mobile-api.md` for server contract.
4. Ensure string resources exist in `values/`, `values-en/`, `values-zh-rCN/`.
5. Run: `cd android && ./gradlew :app:compileDebugKotlin && ./gradlew :app:testDebugUnitTest`

## References

- See `references/runtime-modes.md` for mode behavior details.
- See `docs/mobile-api.md` for the server API contract.
