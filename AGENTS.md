# Agent 操作规范

Commory 是一个本地优先的通信记忆 monorepo。工作时保持上下文收窄，尊重用户已有改动，并验证你实际触碰的范围。

## 先读这些

- Android UI 或移动端产品工作：`android/README.md`、`android/app/src/main/java/com/iskenkenya/commory/mobile/ui/navigation/NavigationHost.kt`、`android/app/src/main/java/com/iskenkenya/commory/mobile/runtime/`，以及目标 screen 或 ViewModel。
- Android 数据、认证、存储工作：`android/sdk/backup`、`android/sdk/auth`、`android/sdk/storage`，以及 `android/app/src/main/java/com/iskenkenya/commory/mobile/remote/`。
- Backend API 工作：`backend/internal/api/server.go`、`backend/internal/auth`、`backend/internal/storage`，以及 `docs/mobile-api.md`。
- Web dashboard 工作：`web/package.json`、`web/src/api`、`web/src/router`，以及目标 view/store 模块。
- MsgLayer/schema 工作：`msglayer/schema/v0.1/root.schema.json`、`msglayer/examples`，以及 `backend/internal/msglayer`。
- 治理或 CI 工作：`docs/engineering-standards.md`、`.github/workflows/ci.yml`、`scripts/`、`.agents/skills`、`.claude/skills`。
- Docker/部署工作：`Dockerfile`、`docker-compose.yml`、`.env.example`、backend config、web production env。

## 上下文包

- Android UI 包：navigation host、目标 screen、`values/`、`values-en/`、`values-zh-rCN/` 中的资源，以及相关 ViewModel。
- Android auth/network 包：runtime environment、auth provider、server client、SDK auth/storage contracts，以及移动端 API 文档。
- Backend API 包：server handler、auth middleware/service、storage provider、API tests，以及移动端 API contract。
- Web dashboard 包：API client、auth store、route guard、目标 view，以及 `web/` 已使用的 Element Plus patterns。
- MsgLayer 包：schema、examples、validators、Android mapper/serializer。
- 治理包：engineering standards、CI workflow、repo hygiene、skills sync、i18n check。
- 部署包：Dockerfile、Compose、环境变量、backend static handler、Web build 输出。

只加载任务需要的上下文包。不要把临时决策扩散到工具专属文件；长期规则放在 `docs/engineering-standards.md`，Agent 入口放在本文件，发布历史放在 `CHANGELOG.md`。

## 结构治理

- 单文件只承载一个 feature slice 或一个清晰的层级职责；不要把多个独立功能继续堆进 `server.go`、`*service.go`、provider 实现文件。
- 新增代码默认先并入现有同职责文件；仅在跨职责、逼近阈值或能明显改善审阅边界时拆分。
- 禁止为了“拆分”制造 1 函数 / 1 类型微文件。新增紧耦合 helper 若总量低于约 `80` 行，默认并入同 feature 文件。
- 手写业务源码进入重构警戒线为 `350` 行，硬上限为 `500` 行；生成代码、schema、fixture、资源文件不纳入硬失败，但超过 `1000` 行仍需在复杂度报告中暴露。
- Web 参考基线以 `web/reference-baseline-paths.txt` 显式登记，覆盖 `references/art-design-pro` 演化而来的路径前缀；这些 baseline 文件会单独报表，不按 Commory-owned Web 代码执行 `500` 行硬失败。
- `bash scripts/check-repo-hygiene.sh` 默认对当前变更集中的 Commory-owned 手写源码执行 `500` 行硬门禁；全仓历史热点与 Web baseline 热区由 `bash scripts/report-loc-complexity.sh` 暴露，并按触碰范围持续治理。
- handler 层按认证、导入、查询、会话、审计等能力分组；auth 层按 orchestration、token、password、audit/rate-limit 分组；storage provider 按 import/query、auth/session、setup/passkey/support 分组；test 按行为域分组，不保留 mega test file。
- 跨 Android、Backend、Web 的改动必须同步评估契约、测试、文档和部署影响。
- 默认按纵向切片推进：先修契约与基础层，再补客户端，再补 UI/治理；不要堆叠不可审阅的大 diff。
- 默认不使用 subagent 或平行代理；只有用户明确要求委派、分工或 parallel agent work 时才可使用。
- 面向用户的最终交付禁止包含原始 scratchpad、逐段回放式研究过程或重复的中间分析；保留决策、变更、验证和风险即可。
- 长期规则变更必须同步更新 `AGENTS.md` 与 `CLAUDE.md`，不能只写进其它文档。

## 仓库边界

- `previewer/` 是历史归档。当前 Commory 工作不要更新它。
- `references/` 是只读外部参考代码。把思路改写进 Commory 自有模块，不要编辑参考镜像。
- `.agents/skills` 是项目 skills 的唯一手工维护来源。
- `.claude/skills` 是 Claude Code 兼容镜像，由脚本生成。不要手改；运行 `bash scripts/sync-agent-skills.sh`。
- 不要新增 `AI_EDIT_LOG.md`、调试报告、已追踪日志、构建产物、`.DS_Store` 或 IDE/cache 文件。

## 项目管理

- `CHANGELOG.md` 记录用户可见变更、兼容性影响和迁移说明；当前版本采用 `v0.x.y`。
- `docs/project-management.md` 定义发布、分支、PR 和两审合并规则。
- `docs/technical-debt.md` 记录已知技术债，不要把无关债务混入功能 PR。
- 禁止直接推送 `main`；`feat/*` 用于功能，`fix/*` 用于修复。
- 跨 Android、Backend、Web、MsgLayer、Agent、Docker 的变更必须同步评估契约、文档、测试和部署影响。
- 大文件或大 diff 若暂时无法继续拆分，必须在 PR 或交付说明里给出简短理由。

## 命名

- Android app id：`com.iskenkenya.commory`。
- Android app namespace 和源码包：`com.iskenkenya.commory.mobile`。
- Android SDK 包：`com.iskenkenya.commory.sdk.*`。
- Go backend module 保持 `github.com/IsKenKenYa/Commory/backend`；它不与 Android 包名冲突。
- 当前用户可见产品名是 `Commory`。历史 changelog 可以保留旧名称作为历史。

## 中文优先与国际化

- 文档、Rules、注释、维护者提示、数据库备注中文优先；必要时第一次出现写作“中文（English）”。
- 产品代码不要为了“中文优先”写死中文。用户可见文本必须进入 i18n/resource 体系。
- Web 使用 `vue-i18n`，语言包位于 `web/src/locales/langs/zh.json` 与 `web/src/locales/langs/en.json`。
- Android 用户可见字符串必须同时维护 `values/strings.xml`、`values-en/strings.xml`、`values-zh-rCN/strings.xml`。
- 系统内置国际化只负责中文和英文；其他语言可以由社区后续扩展。
- 不要翻译代码标识符、JSON key、数据库字段名、HTTP endpoint、配置项、命令、包名、常量值和公开契约。
- `README.en.md` 是英文 README，保持英文；中文 README 更新了产品能力时，同步检查英文 README 是否需要对应更新。

## Web Dashboard

- Web UI 以 `references/art-design-pro` 为模板基线，以 `references/memos/web` 为产品组织参考；只学习结构，不复制参考源码。
- 遇到模糊需求先用 `.agents/skills/grill-me` 收窄边界；改动陌生区域先用 `.agents/skills/zoom-out` 识别影响面；缺陷回归优先按 `.agents/skills/diagnose` / `.agents/skills/tdd` 留下复现与验证；长链路协作或交接前补 `.agents/skills/handoff` 风格摘要。
- 新 Web 页面必须满足中英 i18n、dark mode、主题变量、ECharts 配色、布局密度和 Element Plus 使用规范。
- 具体规则见 `docs/web-dashboard-guidelines.md`。

## 部署

- 默认部署形态是单端口：Go backend 同时提供 `/api/*` 和 Web 静态资源。
- 生产镜像使用根目录 `Dockerfile`；不要新增 `Dockerfile.backend` 或 `web/Dockerfile`，除非后续明确拆分企业部署。
- `docker-compose.yml` 只暴露一个外部端口，默认 `3000:3000`，避免 CORS 成为默认复杂度。
- Web production API 使用同源 `/`，开发环境继续通过 Vite proxy 转发 `/api`。

## Skills

项目 skills 必须遵循 Codex skill 形态：

- 必需：`SKILL.md`，包含简洁 YAML frontmatter（`name`、`description`）和精简 workflow。
- 推荐：需要 UI metadata 时使用 `agents/openai.yaml`。
- 可选：`scripts/`、`references/`、`assets/`。
- 禁止放入 skill：`README.md`、`CHANGELOG.md`、`AGENTS.md`、安装指南、快速参考或大段重复文档。

创建或更新 skill 时，保持 `SKILL.md` 简短，把详细变体移到 `references/` 下直接链接的文件，并运行 `bash scripts/sync-agent-skills.sh`。

## 验证

- Backend：`cd backend && go vet ./... && go test ./... -coverprofile=coverage.out`。
- Android：`cd android && ./gradlew :app:compileDebugKotlin && ./gradlew :app:testDebugUnitTest`。
- Web：`cd web && pnpm install --frozen-lockfile && pnpm lint && pnpm build`。
- 治理：`bash scripts/check-repo-hygiene.sh`、`bash scripts/check-android-i18n.sh`、`bash scripts/sync-agent-skills.sh --check`、`bash scripts/report-loc-complexity.sh`。
- Docker：`docker compose config`、`docker compose build`，必要时 `docker compose up` 后检查 `/` 与 `/api/setup`。
