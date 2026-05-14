[English](README.en.md) | 简体中文

# Commory

> Turn communication into memory.

Commory 是一个通信记忆系统，用来把短信、通话记录、联系人以及后续可扩展的通信数据，转成结构化、可查询、AI-ready 的数据资产。

它不是一个“把文件备份出来就结束”的工具集合。对 Commory 来说，备份只是入口，目标是让个人通信数据变成可以持续使用、分析和构建的基础设施。

## What Is Commory

Commory 面向三个层次的问题：

- 数据采集：从 Android 端把 SMS、Call Log、Contacts 等数据稳定导入
- 数据建模：通过 `MsgLayer` 抽象统一 schema、导出格式和后续 SDK 能力
- 数据使用：提供预览、检索、分析，以及未来的 CLI、插件和 Agent 集成能力

你可以把它理解成：

- `Commory` = 产品层 / 体验层
- `MsgLayer` = 通信数据层 / SDK / schema 能力

## Why It’s Different

传统备份工具的重点是：

- 导出文件
- 保存文件
- 恢复文件

Commory 的重点是：

- 生成结构化数据
- 建立长期可维护的数据层
- 让通信历史可以被检索、分析、组合和复用

换句话说：

- 传统工具产出的是 backup files
- Commory 产出的是 structured, queryable, AI-ready data

## Architecture

```text
Commory
├── Ingestion Layer
│   └── Android app
├── MsgLayer
│   ├── schema
│   ├── export models
│   └── future SDK / CLI interfaces
├── Viewer Layer
│   └── previewer
└── Future Extensions
    ├── CLI
    ├── plugins
    ├── local analysis
    └── agent integrations
```

当前仓库的第一阶段重点，是先把现有 Android 采集能力和 Web 预览能力，收束到 `Commory / MsgLayer` 这套统一叙事之下。

## Components

### `android/`

现有的 Android 入口层，负责设备侧的数据采集、导出、恢复和平台集成。

- 角色：ingestion layer
- 现状：已具备 SMS / 通话记录 / 联系人相关能力
- 技术栈：Kotlin、Jetpack Compose、Material 3、MVVM、多模块 SDK

其中现有 `sdk/backup`、`sdk/auth`、`sdk/storage` 是后续收束到 `MsgLayer` 方向的重要基础。

### `previewer/`

现有的查看层，用于浏览和分析导出的通信数据。

- 角色：viewer layer
- 现状：支持基于 XML 导出内容的本地预览与基础分析
- 技术栈：Vue 3、Vite、Tailwind CSS

`SMS Previewer` 在当前阶段仍然保留为已有组件名称，但不再作为项目主品牌。

## Use Cases

- 搜索自己的通信历史
- 查看联系人互动脉络和时间线
- 对短信和通话记录做本地分析
- 为个人知识库或 Agent 提供结构化通信数据
- 在自托管环境中保留长期可用的数据资产

## Roadmap

第一阶段是对外表达与仓库组织收束；后续路线包括：

- `MsgLayer` 命名与接口继续清晰化
- CLI 能力
- 插件化扩展
- 更多数据源接入
- 本地分析与 AI/Agent 集成
- 统一 viewer 与 mobile 的数据模型

## References

`references/` 用来存放只读参考源码，不参与当前项目构建、发布和许可证主体。

当前包含：

- `references/art-design-pro/`
  - 来源：`https://github.com/Daymychen/art-design-pro.git`
  - 用途：UI、交互和工程组织参考
  - 规则：默认只读，不在本仓直接修改其镜像内容

如果你是贡献者，请不要把 `references/` 下的外部参考代码视为当前项目功能开发目录。

## Current Naming Status

这个仓库正处在品牌与文档收束的第一阶段。

- 首页主品牌现在统一使用 `Commory`
- `MsgLayer` 是正式的数据层架构名
- 仓库历史、源码命名和远程地址中仍可见 `MessageVault`
- `SMS Previewer` 仍作为现有 viewer 组件名称存在

这意味着本次改造只更新对外可见文案，不修改源码包名、Gradle module 名或运行时代码行为。

## Quick Start

### Android

```bash
cd android
./gradlew build
```

### Previewer

```bash
cd previewer
pnpm install
pnpm dev
```

### Clone With Submodules

```bash
git clone <your-repo-url>
cd Commory
git submodule update --init --recursive
```

## GitHub Description

推荐仓库描述：

```text
🧠 Commory · 通信记忆系统｜SMS/Call → Structured Data & AI｜Self-hosted · Privacy-first · Powered by MsgLayer
```

## Contributing

开始开发前请先阅读 [NOTICE.md](NOTICE.md)。

- Android 与 Previewer 保持现有构建方式
- 新的命名收束优先落在文档和架构表达层
- `references/` 为只读参考区，不作为功能实现目录

## License

本仓库根目录代码采用 [GNU General Public License v3.0](LICENSE)。

`references/` 下的外部子模块保持其各自上游仓库的许可证和版权归属，不自动并入本仓主许可证主体。
