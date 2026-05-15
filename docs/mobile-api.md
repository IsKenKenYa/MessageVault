# Mobile API Contract

The Android app uses Commory Server only when the user selects `COMMORY_SERVER`. `LOCAL_ONLY` never requires authentication or network access.

## Envelope

JSON API responses use:

```json
{
  "code": 200,
  "msg": "ok",
  "data": {}
}
```

Errors use the same envelope with `data: null` and an HTTP status code matching `code`. Android displays `msg` as a retryable user-facing error after localization wrapping.

## Setup

- `GET /api/setup`
- Public endpoint.
- Returns whether the server has been initialized.

```json
{
  "status": true,
  "database_type": "sqlite",
  "root_init": true
}
```

## Auth

- `POST /api/auth/register`
- Body: `{ "userName": "alice", "email": "alice@example.com", "password": "..." }`
- Response data: `{ "user": User, "token": "access", "refreshToken": "refresh" }`

- `POST /api/auth/login`
- Body: `{ "userName": "alice", "password": "..." }`
- Response data matches register.

- `POST /api/auth/refresh`
- Body: `{ "refreshToken": "refresh" }`
- Response data: `{ "accessToken": "access", "refreshToken": "refresh" }`

Android sends authenticated requests with `Authorization: Bearer <access token>`.

## User

- `GET /api/user/info`
- Authenticated.
- Returns the current user record. Android uses id, username, email, and roles for session display only.

## Imports

- `GET /api/imports`
- Authenticated.
- Returns import summaries with id, schema version, import timestamp, source path, event count, and identity count.

- `POST /api/imports/upload`
- Authenticated.
- Body: raw `application/json` MsgLayer export or multipart `file`.
- Response data:

```json
{
  "import_id": "import_...",
  "msglayer_version": "msglayer/v0.1"
}
```

- `GET /api/imports/{importId}/export`
- Authenticated.
- Returns the raw MsgLayer JSON export for restore or local inspection.

## Mobile Behavior

- Server mode always writes a local backup first.
- If `syncOnBackup` is enabled and the user is authenticated, Android uploads the generated MsgLayer JSON to `/api/imports/upload`.
- Switching to local mode clears the mobile session but does not delete local backup files.
- Network, auth, and validation failures must not invalidate the local backup.
