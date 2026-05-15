# Android Runtime Modes

## LOCAL_ONLY

- No account or server required.
- Backup and restore use local app storage only.
- Agent context policy: `LOCAL_ONLY` — context must not leave the device.
- Switching into local mode clears auth tokens but preserves local backup files.

## COMMORY_SERVER

- User authenticates with Commory Server (username/password).
- Backup writes local MsgLayer JSON first, then optionally uploads.
- Upload is gated by three conditions: server mode + syncOnBackup + authenticated.
- Remote import history and export use authenticated endpoints.
- Agent context policy: `SERVER_ALLOWED` — future AI features may use server-backed providers with relevance filtering.

## Mode Switching

- Switching modes triggers an `AlertDialog` confirmation.
- Switching to LOCAL_ONLY clears the auth session.
- Switching to COMMORY_SERVER requires server URL and login.
- Local backup files are always preserved across mode switches.

## Session Lifecycle

- `AuthSession.isAuthenticated` checks `accessToken` presence and JWT `exp` claim.
- OkHttp `authenticator` auto-refreshes on 401 (excluding auth endpoints).
- `logoutPersistedSession()` calls `POST /api/auth/logout` with refresh token, then clears local session.
- Refresh token rotation: server issues a new refresh token on each refresh; old one is revoked.
