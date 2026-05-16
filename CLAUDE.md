# CLAUDE.md

Claude Code 应把 `AGENTS.md` 视为项目唯一的规范入口。

## 行为准则（硬约束）

### 编码前思考

- 不要假设。不确定时明确提问，不要默默选一种解释执行。
- 存在多种合理方案时，呈现权衡，不要自行选择。
- 存在更简单的做法时，提出异议。困惑时停下来，指出不清楚的地方。

### 简洁优先

- 用最少的代码解决问题。不要添加要求之外的功能、未请求的抽象或"灵活性"。
- 不为不可能发生的场景做错误处理。200 行能写成 50 行的，重写。
- 检验标准：资深工程师会觉得过于复杂吗？如果是，简化。

### 精准修改

- 只碰必须碰的。不要"改进"相邻代码、注释或格式，不要重构没坏的东西。
- 匹配现有风格，即使你偏好不同写法。
- 注意到无关死代码时提一下，不要自行删除。
- 你的改动产生的孤儿代码（无用的 import/变量/函数）由你清理；预先存在的死代码不动。
- 检验标准：每一行修改都应能直接追溯到用户的请求。

### 目标驱动执行

- 将指令式任务转化为可验证目标："添加验证" → "为无效输入写测试，然后让它们通过"。
- 多步骤任务先列简短计划，每步附验证方式。
- 强成功标准让 Agent 能独立循环；弱标准（"让它工作"）需要不断澄清。


**权衡**：这些准则倾向于谨慎而非速度。琐碎任务（显而易见的一行修改）自行判断，不必走完整流程。目标是减少非琐碎工作中的代价高昂的错误。

## 结构治理

- 单文件只承载一个 feature slice 或一个清晰职责，不继续向 `server.go`、`*service.go`、provider 实现文件堆多个独立功能。
- 新增代码默认先并入现有同职责文件；仅在跨职责、逼近阈值或能明显改善审阅边界时拆分。
- 禁止为了"拆分"制造 1 函数 / 1 类型微文件；低于约 `80` 行的紧耦合 helper 默认并入同 feature 文件。
- 手写业务源码 `350` 行进入重构警戒线，`500` 行为硬上限；生成代码、schema、fixture、资源文件不做硬失败，但要在复杂度报告里暴露 `1000+` 文件。
- Web 参考基线通过 `web/reference-baseline-paths.txt` 显式维护；这些参考衍生文件单独报表，不按 Commory-owned Web 代码执行 `500` 行硬失败。
- `bash scripts/check-repo-hygiene.sh` 默认拦截当前变更集中的 Commory-owned `500+` 手写源码；全仓历史热点和 Web baseline 热区继续通过 `bash scripts/report-loc-complexity.sh` 报告。
- handler 层按认证、导入、查询、会话、审计等能力分组；auth 层按 orchestration、token、password、audit/rate-limit 分组；storage provider 按 import/query、auth/session、setup/passkey/support 分组。
- 跨 Android、Backend、Web 的改动必须同步评估契约、测试、文档和部署影响，并按纵向切片推进，避免不可审阅的大 diff。
- 默认不使用 subagent 或平行代理；只有用户明确要求委派或分工时才可使用。
- 面向用户的交付不要回放原始 scratchpad 或重复中间研究结论，只保留决策、变更、验证和风险。
- 长期规则变更必须同步更新 `AGENTS.md` 与 `CLAUDE.md`；详细说明服从 `docs/engineering-standards.md`。

## 仓库边界

- `previewer/` 是历史归档。当前 Commory 工作不要更新它。
- `references/` 是只读外部参考代码。把思路改写进 Commory 自有模块，不要编辑参考镜像。
- `.agents/skills` 是项目 skills 的唯一手工维护来源。
- `.claude/skills` 是 Claude Code 兼容镜像，由脚本生成。不要手改；运行 `bash scripts/sync-agent-skills.sh`。
- 不要新增 `AI_EDIT_LOG.md`、调试报告、已追踪日志、构建产物、`.DS_Store` 或 IDE/cache 文件。

## 项目管理

- `CHANGELOG.md` 记录用户可见变更、兼容性影响和迁移说明。
- `docs/project-management.md` 定义发布、分支、PR 和两审合并规则。
- `docs/technical-debt.md` 记录已知技术债，不要把无关债务混入功能 PR。
- 禁止直接推送 `main`；`feat/*` 用于功能，`fix/*` 用于修复。

## 中文优先与国际化

- 文档、Rules、注释、维护者提示、数据库备注中文优先。
- 产品代码不要为了"中文优先"写死中文。用户可见文本必须进入 i18n/resource 体系。
- Web 使用 `vue-i18n`，语言包位于 `web/src/locales/langs/zh.json` 与 `web/src/locales/langs/en.json`。
- Android 用户可见字符串必须同时维护 `values/strings.xml`、`values-en/strings.xml`、`values-zh-rCN/strings.xml`。
- 不要翻译代码标识符、JSON key、数据库字段名、HTTP endpoint、配置项、命令、包名、常量值和公开契约。
- `README.en.md` 保持英文，中文 README 是主要维护入口。

## Skills

- `.agents/skills` 是项目 skills 的唯一手工维护来源。
- `.claude/skills` 是 Claude Code 兼容镜像。
- 不要直接编辑 `.claude/skills`。先更新 `.agents/skills`，再运行：

```bash
bash scripts/sync-agent-skills.sh
```

创建或更新 skill 时，保持 `SKILL.md` 简短，把详细变体移到 `references/` 下直接链接的文件。

## 部署与 Web

- 默认部署是单端口：Go backend 同时提供 `/api/*` 和 Web 静态资源。
- 根目录 `Dockerfile` 是完整 Commory 镜像；`docker-compose.yml` 默认只暴露 `3000`。
- `docker-compose.yml` 只暴露一个外部端口，默认 `3000:3000`，避免 CORS 成为默认复杂度。
- Web production API 使用同源 `/`，开发环境继续通过 Vite proxy 转发 `/api`。
- Web UI 必须遵循 `docs/web-dashboard-guidelines.md`，以 `references/art-design-pro` 为模板基线。
- 模糊需求先跑 `.agents/skills/grill-me`，陌生区域先做 `.agents/skills/zoom-out`，回归问题优先按 `.agents/skills/diagnose` 或 `.agents/skills/tdd` 留痕。
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
