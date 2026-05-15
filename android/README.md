# Commory Android

Commory Android is the local-first client for collecting, exporting, restoring, and optionally syncing communication data.

## Identity

- Application id: `com.iskenkenya.commory`
- App namespace: `com.iskenkenya.commory.mobile`
- SDK namespaces: `com.iskenkenya.commory.sdk.backup`, `com.iskenkenya.commory.sdk.auth`, `com.iskenkenya.commory.sdk.storage`
- Product name: `Commory`

The Go backend module name is independent from Android package naming and does not conflict with the client id.

## Runtime Modes

- `LOCAL_ONLY`: backup and restore run on device without auth or network.
- `COMMORY_SERVER`: local backup still happens first, then authenticated server upload can run when enabled.

Mode behavior is documented in [../docs/android-runtime-modes.md](../docs/android-runtime-modes.md).

## Modules

```text
android/
├── app/          # Compose UI, navigation, permissions, runtime mode, Android platform adapters
└── sdk/
    ├── backup/  # Pure Kotlin backup/restore orchestration and MsgLayer mapping
    ├── auth/    # Pure Kotlin auth contracts
    └── storage/ # Android storage implementation and storage contracts
```

`app` may depend on SDK modules. SDK modules must not depend on `app`; pure Kotlin SDK modules must not import Android framework APIs.

## Commands

```bash
./gradlew :app:compileDebugKotlin
./gradlew :app:testDebugUnitTest
./gradlew :sdk:backup:test
./gradlew :sdk:auth:test
```

Run Android checks from the `android/` directory. Use instrumented tests only when a device or emulator is required.

## Documentation

- Engineering standards: [../docs/engineering-standards.md](../docs/engineering-standards.md)
- Mobile API contract: [../docs/mobile-api.md](../docs/mobile-api.md)
- Runtime modes: [../docs/android-runtime-modes.md](../docs/android-runtime-modes.md)
- Changelog: [CHANGELOG.md](CHANGELOG.md)

Historical reports such as `AI_EDIT_LOG.md` are not used. Durable changes belong in `CHANGELOG.md`, PR descriptions, and commits.
