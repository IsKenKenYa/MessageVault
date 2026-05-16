# CLAUDE.md

Claude Code 应把 `AGENTS.md` 视为项目唯一的规范入口。

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
