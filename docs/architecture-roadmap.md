# Commory 架构路线图

Commory 正在从原型存储和直接 handler 走向持久、可测试的 service 边界。近期工作刻意保持朴素：先补数据库正确性、认证加固和架构护栏，再扩展更丰富的 Agent 行为。

## 阶段 1：持久化基础

- 把文件存储伪装的 `sqlite` 和 `postgres` provider alias 替换为真实 SQL 实现。
- 使用 `sqlc` 加版本化 migrations 支持 SQLite、Postgres 和 MySQL。
- 为 auth、setup、import/export、search、timeline、identities 和 refresh token rotation 增加 storage contract tests。
- setup diagnostics 必须明确当前 database type 和 SQLite persistence risk。
- Docker Compose 先使用 file-backed sqlite 开发部署，后续 SQL 迁移再扩展 Postgres/MySQL profile。

## 阶段 2：认证与权限

- 保持短生命周期 access tokens，并使用 refresh token rotation。
- 增加 session/device records、revoke audit，以及 user/admin/root 行为的 typed role constants。
- HTTP middleware 保持薄层；认证授权决策放在 service-level code。

## 阶段 3：Agent Runtime

- 在 `backend/internal/agent` 建立 Go 自有 runtime boundary。
- 通过 interface 增加 provider adapters。
- 在允许 tool 修改本地状态前，先建立 tool registry 和 permission broker。
- 通过 storage interfaces 持久化 task/session state。

## 阶段 4：Web Dashboard 与单端口部署

- Go backend 同源提供 `/api/*` 与 Web 静态资源，生产默认避免 CORS。
- Web 以 Art Design Pro 为模板基线，所有新增页面满足 i18n、dark mode、主题变量和图表配色规范。
- Dashboard 功能按 Android、Backend、Web、MsgLayer 同步规划，避免只做单端体验。

## 阶段 5：参考资料挖掘

- 使用 GitHubDaily 发现候选 Go 项目。
- 优先选择 MIT、Apache-2.0 和 BSD 参考项目。
- AGPL/GPL/LGPL/unknown-license 项目只用于只读架构观察。
- 每个被接受的参考项目都必须记录在 `references/README.md`。
