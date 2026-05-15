# 工程标准

这是 Commory 的唯一工程标准。长期规则放在这里，Agent 工作流入口放在 `AGENTS.md`，发布历史放在 `CHANGELOG.md`。

## 架构边界

- Android `app` 负责 Compose UI、权限、平台读写器、导航、运行模式设置和 Android 专属网络能力。
- Android `sdk/backup` 负责纯 Kotlin 备份/恢复编排和 MsgLayer 映射。
- Android `sdk/auth` 负责纯 Kotlin 认证契约。
- Android `sdk/storage` 负责存储契约和 Android 存储实现。
- Backend handler 保持薄层：解析 HTTP、调用 auth/storage/query/import/setup service、返回 envelope。
- Web 负责 server-backed dashboard workflow，并遵循 `web/` 现有 Vue 3、Vite、Pinia、Element Plus 模式。
- MsgLayer schema 是 Android、backend、web 和未来 agent 之间的 canonical interchange format。

## 命名

- Android application id：`com.iskenkenya.commory`。
- Android app namespace 和源码包：`com.iskenkenya.commory.mobile`。
- Android SDK 包：`com.iskenkenya.commory.sdk.*`。
- Backend Go module 保持 `github.com/IsKenKenYa/Commory/backend`。
- 当前用户可见产品名是 `Commory`；旧名称只能作为历史保留在历史 changelog 中。

## 仓库边界

- `previewer/` 是历史归档。当前 roadmap、CI、文档或命名工作不要更新它。
- `references/` 是只读外部参考代码。
- `.agents/skills` 是项目 skills 的唯一手工维护来源。
- `.claude/skills` 由 `.agents/skills` 生成，用于 Claude Code 兼容。
- 不要新增已追踪构建产物、日志、`.DS_Store`、IDE 文件、调试报告或 `AI_EDIT_LOG.md`。

## 中文优先与国际化

- 文档、Rules、代码注释、维护者提示、测试失败信息和数据库备注中文优先。
- 产品代码里的用户可见文本必须走国际化，不要因为中文优先而硬编码中文。
- Web 使用 `vue-i18n`，配置在 `web/src/locales/index.ts`，内置语言包为 `web/src/locales/langs/zh.json` 和 `web/src/locales/langs/en.json`。
- Android 用户可见字符串必须同时存在于 `values/strings.xml`、`values-en/strings.xml`、`values-zh-rCN/strings.xml`。
- Compose UI 必须使用 `stringResource`。
- 动态错误应在 UI 边界前类型化，并通过占位符本地化。
- 系统内置国际化只负责中文和英文；其它语言留给社区扩展。
- 不要翻译代码标识符、JSON key、数据库字段名、HTTP endpoint、配置项、命令、包名、常量值和公开契约。
- `README.md` 是中文优先入口；`README.en.md` 保持英文镜像。

## Skills

Skills 是上下文加载工具，不是通用文档目录。

- 每个 skill 必需：`SKILL.md`，包含简洁 frontmatter（`name`、`description`）和精简 workflow。
- 推荐：需要 UI metadata 时使用 `agents/openai.yaml`。
- 可选：`scripts/`、`references/`、`assets/`。
- skill 内禁止：`README.md`、`CHANGELOG.md`、`AGENTS.md`、安装指南、快速参考和重复的大段文档。

更新 skills 时，编辑 `.agents/skills`，然后运行：

```bash
bash scripts/sync-agent-skills.sh
```

CI 或 review 中使用 `bash scripts/sync-agent-skills.sh --check` 确认 `.claude/skills` 没有漂移。

## 上下文管理

使用能完成任务的最小上下文包：

- Android UI：navigation host、目标 screen、字符串资源、相关 ViewModel。
- Android auth/network：runtime environment、auth provider、server client、SDK contracts、mobile API docs。
- Backend API：server handler、auth middleware/service、storage provider、API tests、mobile API docs。
- Web：API client、auth store、route guards、目标 view、现有 dashboard conventions。
- MsgLayer：schema、examples、validators、Android mapper/serializer。
- 治理：本文件、`AGENTS.md`、CI workflow、scripts、skills。

除非能消除真实重复或澄清跨模块契约，否则不要新增共享抽象。大文件或大 diff 需要拆分，或在 PR 中给出简短理由。

## 日志与隐私

- 日志可以标识 subsystem 和 operation context。
- 日志不得包含短信正文、联系人内容、auth token、refresh token、备份载荷，或带凭据的私有服务器 URL。
- 本地备份已经成功时，服务器上传失败必须是非破坏性的。
- Local-only mode 不得要求 backend auth 或网络。

## 测试

- Backend 核心包应重点覆盖 storage、auth、validators、import/query logic、setup 和 mobile contracts。
- Android SDK 纯 Kotlin 逻辑使用 JVM tests。
- Android ViewModel/domain 行为尽量使用 JVM tests。
- Instrumented tests 只用于权限、content providers、Compose flows 和平台行为。
- Web 变更必须通过 lint 和 production build。

## 质量门禁

```bash
cd backend && go vet ./... && go test ./... -coverprofile=coverage.out
cd android && ./gradlew :app:compileDebugKotlin && ./gradlew :app:testDebugUnitTest
cd web && pnpm install --frozen-lockfile && pnpm lint && pnpm build
bash scripts/check-repo-hygiene.sh
bash scripts/check-android-i18n.sh
bash scripts/sync-agent-skills.sh --check
```

## 发布检查

- Mobile API contract 仍与 `docs/mobile-api.md` 一致。
- Local-only mode 在没有网络或 auth 时可用。
- Server mode 覆盖 setup、login/register、refresh、logout、upload、list、export flows。
- Android locale keys 完整。
- 跨边界 API/schema 变更包含文档和测试。
- Changelog 说明用户可见变更和兼容性变化。
