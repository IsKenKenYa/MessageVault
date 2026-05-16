# CLAUDE.md

Claude Code 应把 `AGENTS.md` 视为项目唯一的规范入口。

## 结构治理

- 单文件只承载一个 feature slice 或一个清晰职责，不继续向 `server.go`、`*service.go`、provider 实现文件堆多个独立功能。
- 新增代码默认先并入现有同职责文件；仅在跨职责、逼近阈值或能明显改善审阅边界时拆分。
- 禁止为了“拆分”制造 1 函数 / 1 类型微文件；低于约 `80` 行的紧耦合 helper 默认并入同 feature 文件。
- 手写业务源码 `350` 行进入重构警戒线，`500` 行为硬上限；生成代码、schema、fixture、资源文件不做硬失败，但要在复杂度报告里暴露 `1000+` 文件。
- `bash scripts/check-repo-hygiene.sh` 默认拦截当前变更集中的 `500+` 手写源码；全仓历史热点继续通过 `bash scripts/report-loc-complexity.sh` 报告。
- 跨 Android、Backend、Web 的改动必须同步评估契约、测试、文档和部署影响，并按纵向切片推进，避免不可审阅的大 diff。
- 默认不使用 subagent 或平行代理；只有用户明确要求委派或分工时才可使用。
- 面向用户的交付不要回放原始 scratchpad 或重复中间研究结论，只保留决策、变更、验证和风险。
- 长期规则变更必须同步更新 `AGENTS.md` 与 `CLAUDE.md`；详细说明服从 `docs/engineering-standards.md`。

## Skills

- `.agents/skills` 是项目 skills 的唯一手工维护来源。
- `.claude/skills` 是 Claude Code 兼容镜像。
- 不要直接编辑 `.claude/skills`。先更新 `.agents/skills`，再运行：

```bash
bash scripts/sync-agent-skills.sh
```

## 中文优先与国际化

- 文档、Rules、注释、维护者提示和数据库备注中文优先。
- 代码标识符、命令、路径、API 字段、协议值和公开契约保持英文/原样。
- 产品中的用户可见文案必须走 i18n/resource，不要在代码里硬编码中文。
- `README.en.md` 保持英文，中文 README 是主要维护入口。

## 部署与 Web

- 默认部署是单端口：Go backend 同时提供 `/api/*` 和 Web 静态资源。
- 根目录 `Dockerfile` 是完整 Commory 镜像；`docker-compose.yml` 默认只暴露 `3000`。
- Web UI 必须遵循 `docs/web-dashboard-guidelines.md`，以 `references/art-design-pro` 为模板基线。
- 多端功能必须同步评估 Android、Backend、Web、MsgLayer、Agent 和 Docker/部署影响。

## 常用命令

```bash
bash scripts/dev.sh                    # 一键启动 backend (:3000) + web (:3006)
bash scripts/check-repo-hygiene.sh
bash scripts/check-android-i18n.sh
bash scripts/sync-agent-skills.sh --check
docker compose config
```

代码工作使用 `AGENTS.md` 中列出的验证命令。
