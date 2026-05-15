[English](README.en.md) | 简体中文

# Commory

> Turn communication into memory.

Commory 是一个本地优先的通信记忆系统，用来把短信、通话记录、联系人以及后续可扩展的数据源转成结构化、可查询、可同步、AI-ready 的个人数据资产。

Commory 不是“导出一个备份文件就结束”的工具。备份是入口，目标是让个人通信历史成为长期可维护的数据层。

## 架构

```text
Commory
├── android/     # Android 客户端：本地备份、恢复、可选服务器同步
├── backend/     # Commory Server：认证、导入、查询、自托管 API
├── web/         # Vue dashboard：服务器管理与查看体验
├── msglayer/    # Canonical schema：跨端通信数据契约
├── docs/        # 当前工程标准与 API 文档
├── scripts/     # CI 与治理脚本
├── previewer/   # 历史 XML SMS viewer 归档，不再更新
└── references/  # 只读外部参考源码
```

当前状态是“本地备份 + 可选服务器上传”。下一步演进目标是双向同步：上传、远程恢复、增量同步。远期方向是端到端加密、多设备同步和 AI Agent 上下文。

核心原则保持不变：本地优先，服务器可选，隐私可控。

## 中文优先与国际化

中文是项目维护资料的优先语言，便于当前维护者快速理解架构、规则和上下文；[README.en.md](README.en.md) 保持英文版，不会因为中文优先而放弃国际化。

产品代码中的用户可见文案不应硬编码中文或英文。Web 使用 `vue-i18n`，内置 `web/src/locales/langs/zh.json` 与 `web/src/locales/langs/en.json`；Android 使用 `values/`、`values-en/`、`values-zh-rCN/` 资源。系统内置国际化只负责中文和英文，更多语言交给社区扩展。

## 组件

### `android/`

Commory Android 客户端，负责设备侧数据采集、MsgLayer 导出、本地恢复、运行模式选择和可选服务器同步。

- 包名：`com.iskenkenya.commory`
- app namespace：`com.iskenkenya.commory.mobile`
- SDK namespace：`com.iskenkenya.commory.sdk.*`
- 技术栈：Kotlin、Jetpack Compose、Material 3、DataStore、Retrofit、多模块 SDK

### `backend/`

自托管 Commory Server，负责认证、Refresh Token、MsgLayer 导入、查询、时间线、搜索和移动端 API。

- 技术栈：Go、标准库 HTTP、SQLite/PostgreSQL provider
- 契约文档：[docs/mobile-api.md](docs/mobile-api.md)

### `web/`

服务器端 dashboard，基于 Vue 3、Vite、Element Plus，用于管理与查看 server-backed 数据。

### `msglayer/`

Commory 的 canonical interchange format。Android 输出 MsgLayer JSON，backend 验证并导入，未来 CLI/SDK/Agent 能力也围绕 MsgLayer 演进。

### `previewer/`

历史 XML SMS viewer 归档。它保留作为项目早期代码存档，不参与当前构建、CI、规范迁移或功能路线。

## Quick Start

### Android

```bash
cd android
./gradlew :app:compileDebugKotlin
./gradlew :app:testDebugUnitTest
```

### Backend

```bash
cd backend
go test ./...
go run ./cmd/commory
```

### Web

```bash
cd web
pnpm install
pnpm dev
```

## Governance

- Agent 入口：[AGENTS.md](AGENTS.md)
- 工程标准：[docs/engineering-standards.md](docs/engineering-standards.md)
- 项目治理：[NOTICE.md](NOTICE.md)
- 当前移动端契约：[docs/mobile-api.md](docs/mobile-api.md)
- ADR：
  - [参考资料治理](docs/adr/0001-参考资料治理.md)
  - [数据库 sqlc 迁移](docs/adr/0002-数据库-sqlc-迁移.md)
  - [Agent 运行时边界](docs/adr/0003-agent-运行时边界.md)

`.agents/skills` 是项目 Skills 的唯一手工维护来源，`.claude/skills` 是 Claude Code 兼容镜像。更新 Skills 后运行：

```bash
bash scripts/sync-agent-skills.sh
```

## Roadmap

- 阶段 1：安全与健壮性补全，包括 Android logout 调服务端、401 自动刷新、去除阻塞快照、CI Java 25、治理规范收束。
- 阶段 2：移动端体验完善，包括网络错误分类、token 过期判断、模式切换确认、服务器地址默认策略、Retrofit 缓存。
- 阶段 3：契约测试补全，包括 refresh 轮换、logout 吊销、setup 初始化防重、web lint/build CI。
- 阶段 4：功能扩展，包括远程恢复、搜索与时间线、推送通知、MsgLayer v0.2、端到端加密、多设备同步。

## License

本仓库根目录代码采用 [GNU General Public License v3.0](LICENSE)。`references/` 下外部参考代码保持其上游许可证。
