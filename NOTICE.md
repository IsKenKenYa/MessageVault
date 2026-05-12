# MessageVault 项目规范

## 🏗️ 编码标准

### 命名规范
- **Kotlin**: 遵循 [Kotlin 编码规范](https://kotlinlang.org/docs/coding-conventions.html)
- **JavaScript/Vue**: 遵循 [Vue.js 风格指南](https://vuejs.org/style-guide/)
- **文件命名**: Kotlin 使用 PascalCase（类文件）和 camelCase（工具文件），Vue 使用 PascalCase

### 代码风格
- 使用一致的缩进（Kotlin 4空格，JavaScript 2空格）
- 每个文件末尾保留一个空行
- 避免过长的行（建议 120 字符以内）

### 架构原则
- Android: MVVM + 多模块SDK架构
- Vue: Composition API + 单向数据流
- 模块化设计，职责分离
- SDK模块接口驱动，平台实现通过构造函数注入

## 🧩 多模块开发规范

### 模块依赖规则

```
app → sdk/backup, sdk/auth, sdk/storage
sdk/backup ✗→ sdk/auth, sdk/storage, app
sdk/auth   ✗→ sdk/backup, sdk/storage, app
sdk/storage ✗→ sdk/backup, sdk/auth, app
```

- `app` 模块可以依赖所有 `sdk/*` 模块
- `sdk/*` 模块之间**禁止互相依赖**
- `sdk/*` 模块**禁止依赖** `app` 模块
- 新增模块依赖关系需经架构评审

### SDK模块编码标准

#### 纯Kotlin模块（sdk/backup、sdk/auth）

- **禁止引入Android依赖**：不得使用 `android.*`、`androidx.*` 等Android框架API
- **构建类型**：使用 `java-library` + `org.jetbrains.kotlin.jvm` 插件
- **依赖限制**：仅允许纯Kotlin/JVM依赖（kotlin-stdlib、kotlinx-coroutines-core、gson等）
- **接口设计**：所有平台相关操作通过接口暴露（如 `SmsReader`、`AuthProvider`），由 `app` 模块提供Android实现
- **可测试性**：所有逻辑可通过纯JVM单元测试验证，无需Android模拟器

#### Android Library模块（sdk/storage）

- **构建类型**：使用 `com.android.library` + `org.jetbrains.kotlin.android` 插件
- **允许的Android依赖**：Room、Retrofit、androidx.core等AndroidX库
- **接口设计**：对外暴露纯Kotlin接口（如 `StorageProvider`），内部使用Android API实现
- **最小化Android耦合**：尽量将纯逻辑提取为独立函数，减少对Android框架的依赖

#### App模块

- **职责**：UI展示、导航、权限管理、Android平台SDK接口实现
- **SDK接口实现**：所有SDK定义的接口（`SmsReader`、`SmsWriter`、`StorageProvider`等）在 `app` 模块中提供Android实现
- **依赖注入**：通过构造函数注入SDK接口实现，不使用Service Locator或全局单例

### 测试规范

| 模块类型 | 测试方式 | 测试命令 | 说明 |
|---------|---------|---------|------|
| 纯Kotlin SDK | JVM单元测试 | `./gradlew :sdk:backup:test` | 无需Android设备/模拟器 |
| 纯Kotlin SDK | JVM单元测试 | `./gradlew :sdk:auth:test` | 无需Android设备/模拟器 |
| Android Library | Instrumented测试 | `./gradlew :sdk/storage:connectedAndroidTest` | 需要Android设备/模拟器 |
| App | JVM + Instrumented | `./gradlew :app:test` | ViewModel用JVM测试，UI用Instrumented测试 |

- SDK模块测试覆盖率目标：核心逻辑 > 80%
- App模块测试覆盖率目标：ViewModel > 70%
- 新增SDK接口必须附带接口契约测试

## 💬 注释要求

- 公共 API 必须包含文档注释
- 复杂逻辑必须添加行内注释
- 注释使用中文或英文，保持项目内一致

## 🔧 可扩展性

- 使用接口/抽象类定义扩展点
- 遵循开闭原则（对扩展开放，对修改关闭）
- 新功能应通过模块方式添加

## 🧪 测试要求

- 核心功能必须有单元测试
- 关键流程需要集成测试
- 测试覆盖率目标：SDK核心模块 > 80%，App ViewModel > 70%
- 纯Kotlin SDK模块（sdk/backup、sdk/auth）使用JVM单元测试，无需Android模拟器
- Android Library模块（sdk/storage）使用Instrumented测试
- 新增SDK接口必须附带接口契约测试

## 📝 变更记录

- 每个版本更新必须更新 CHANGELOG.md
- 遵循 [Keep a Changelog](https://keepachangelog.com/) 格式
- AI 辅助编辑需记录到 AI_EDIT_LOG.md

## 📚 文档要求

- 每个子项目必须有 README.md
- API 变更必须更新文档
- 用户可见的变更需更新用户手册

## 📦 Monorepo 结构

```
MessageVault/
├── android/                    # MessageVault-Mobile (Kotlin/Android)
│   ├── sdk/
│   │   ├── backup/             # 纯Kotlin备份/恢复SDK
│   │   ├── auth/               # 纯Kotlin认证组件
│   │   └── storage/            # Android Library存储组件
│   ├── app/                    # Android应用壳
│   └── docs/architecture/      # 架构设计文档
│       ├── harmonyos-adaptation.md
│       ├── third-party-auth.md
│       ├── ai-agent-integration.md
│       └── backend-microservices.md
├── previewer/                  # SMS-Previewer (Vue 3/Vite)
├── .trae/                      # 开发工具配置
├── LICENSE                     # GPL v3.0
├── NOTICE.md                   # 项目规范
└── README.md                   # 项目总览
```
