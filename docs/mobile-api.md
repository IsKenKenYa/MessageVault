# 移动端 API 契约

Android app 只有在用户选择 `COMMORY_SERVER` 时才使用 Commory Server。`LOCAL_ONLY` 从不要求认证或网络访问。

## Envelope

JSON API 响应使用：

```json
{
  "code": 200,
  "msg": "ok",
  "data": {}
}
```

错误响应使用相同 envelope，`data: null`，HTTP status code 与 `code` 对齐。Android 不应直接把服务端 `msg` 当最终文案展示；它应包裹为可重试、可本地化的用户错误。

## Setup

- `GET /api/setup`
- 公开 endpoint。
- 返回服务器是否已经初始化。

```json
{
  "status": true,
  "database_type": "sqlite",
  "root_init": true
}
```

## Auth

- `POST /api/auth/register`
- Body：`{ "userName": "alice", "email": "alice@example.com", "password": "..." }`
- Response data：`{ "user": User, "token": "access", "refreshToken": "refresh" }`

- `POST /api/auth/login`
- Body：`{ "userName": "alice", "password": "..." }`
- Response data 与 register 一致。

- `POST /api/auth/refresh`
- Body：`{ "refreshToken": "refresh" }`
- Response data：`{ "token": "access", "refreshToken": "refresh" }`

Android 通过 `Authorization: Bearer <access token>` 发送认证请求。

## User

- `GET /api/user/info`
- 需要认证。
- 返回当前用户记录。Android 只使用 id、username、email 和 roles 做 session 展示。

## Imports

- `GET /api/imports`
- 需要认证。
- 返回 import summaries，包括 id、schema version、import timestamp、source path、event count 和 identity count。

- `POST /api/imports/upload`
- 需要认证。
- Body：原始 `application/json` MsgLayer export 或 multipart `file`。
- Response data：

```json
{
  "import_id": "import_...",
  "msglayer_version": "msglayer/v0.1"
}
```

- `GET /api/imports/{importId}/export`
- 需要认证。
- 返回原始 MsgLayer JSON export，用于恢复或本地检查。

## 移动端行为

- Server mode 始终先写本地备份。
- `syncOnBackup` 启用且用户已认证时，Android 把生成的 MsgLayer JSON 上传到 `/api/imports/upload`。
- 切回 local mode 会清除 mobile session，但不删除本地备份文件。
- 网络、认证和验证失败不得让已经生成的本地备份失效。
