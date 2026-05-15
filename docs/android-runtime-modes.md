# Android Runtime Modes

Commory Android supports two explicit runtime modes.

## Local Only

- No account or server is required.
- Backup and restore use local app storage.
- Agent context policy is `LOCAL_ONLY`; context must not be sent to server-backed providers.
- Switching into local mode clears mobile auth tokens but preserves local backup files.

## Commory Server

- User signs in with the existing Commory username/password auth flow.
- Backup still writes a local MsgLayer JSON file first.
- When `syncOnBackup` is enabled and a valid session exists, Android uploads the generated file to `/api/imports/upload`.
- Remote import history and export use authenticated Commory Server endpoints.
- Agent context policy is `SERVER_ALLOWED`, but future AI features must still apply relevance and minimum-necessary context filtering.
