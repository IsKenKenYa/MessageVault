# 项目管理规范

## 发布管理

- 使用 `CHANGELOG.md` 记录每个里程碑的用户可见变更、兼容性影响和迁移说明。
- 版本采用 `v0.x.y` 格式。当前处于 `v0` 阶段，API、MsgLayer schema 和部署方式仍可能变化。
- 每次发布前逐项确认 `docs/engineering-standards.md` 中的 Release Checklist。

## 代码审查流程

- 所有非文档小改都通过 PR 合并，PR 使用 `.github/pull_request_template.md`。
- PR 必须说明变更描述、影响范围、测试方法、契约一致性检查、i18n、隐私、本地模式和 Docker 影响。
- 涉及 Android + Backend + Web + MsgLayer 任意两个以上模块的 PR，至少需要两个审查者。
- 审查时必须检查 Commory 特定项：local-only mode、server-optional、隐私日志、Agent context policy、API/MsgLayer 契约和中英国际化。

## 分支策略

- `main` 是稳定分支，要求 CI 全绿。
- `feat/*` 用于功能分支。
- `fix/*` 用于修复分支。
- 禁止直接推送到 `main`。该规则需要在 GitHub branch protection 中启用。

## 多端规划

- 跨端功能必须同时评估 Android、Backend、Web、MsgLayer、Agent 和 Docker/部署影响。
- API 或 schema 先更新契约文档，再改实现。
- Web 功能必须遵循 `docs/web-dashboard-guidelines.md`。
- 部署相关功能优先考虑单端口同源方案，避免 CORS 成为默认复杂度。

## 文档维护

- `docs/mobile-api.md` 是 Android 和 backend 的契约文档，任何 API 变更必须同步更新。
- `docs/engineering-standards.md` 是工程规范唯一真相源。
- `AGENTS.md` 是 Agent 入口文件，保持精简。
- `.agents/skills` 是项目知识库来源；修改后必须运行 `bash scripts/sync-agent-skills.sh`。
