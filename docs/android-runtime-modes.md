# Android 运行模式

Commory Android 支持两个明确的运行模式。

## Local Only

- 不需要账号或服务器。
- 备份和恢复使用本地 app storage。
- Agent context policy 为 `LOCAL_ONLY`；上下文不得发送给 server-backed provider。
- 切换到 local mode 会清除移动端 auth tokens，但保留本地备份文件。

## Commory Server

- 用户使用现有 Commory username/password auth flow 登录。
- 备份仍然先写入本地 MsgLayer JSON 文件。
- `syncOnBackup` 启用且存在有效 session 时，Android 把生成文件上传到 `/api/imports/upload`。
- 远程 import history 和 export 使用已认证的 Commory Server endpoints。
- Agent context policy 为 `SERVER_ALLOWED`，但未来 AI 功能仍必须应用相关性过滤和最小必要上下文过滤。
