# Commory Android 功能需求

本文档列出了Commory Android应用的所有核心功能和实现原则，以指导开发流程和优先级排序。

## 当前开发状态

**重要提示：** 本项目目前处于积极开发阶段，功能尚未完全实现。请注意以下事项：

1. **测试风险警告**：
   - 在测试前请务必使用设备生产商自带的备份还原软件备份您的短信、通话记录和联系人
   - 本软件可能会出现不可预见的错误，可能导致数据丢失
   - 建议仅使用备份功能，恢复功能请在虚拟机环境中测试

2. **功能完成度**：
   - 备份功能：基本可用，支持短信、通话记录和联系人备份
   - 恢复功能：已支持通话记录和短信恢复，联系人恢复功能尚未完全实现
   - 界面：基本完成，使用Material Design 3实现
   - 默认短信应用检测：优化完成，支持Android 7.0-14，解决了高版本Android上的检测问题

3. **开发环境**：
   - JDK版本：JDK 17 (OpenJDK 17.0.14)
   - Gradle版本：8.2
   - Android SDK：API 24-34
   - 编译目标：Android 7.0 (API 24) 及以上

## 核心功能

### 1. 数据访问与读取

- **短信读取**：
  - 读取设备上的所有SMS消息，包括收件箱和发件箱
  - 支持分页加载以优化性能
  - 保存关键元数据（发送者/接收者、日期、内容、类型）
  
- **通话记录读取**：
  - 读取设备通话历史记录
  - 包括来电、去电和未接来电
  - 记录通话时间、持续时间和号码

- **内容提供者访问**：
  - 使用ContentResolver安全访问系统数据
  - 遵循Android权限最佳实践
  - 适配不同Android版本（API 21-34）的数据访问差异

### 2. 备份功能

- **本地备份**：
  - 将数据导出为JSON格式
  - 保存到设备外部存储（遵循分区存储规则）
  - 使用时间戳命名备份文件
  - 提供备份摘要和统计信息

- **远程备份**：
  - 通过HTTPS安全上传数据到服务器
  - 支持断点续传和重试机制
  - 备份进度显示和状态通知
  - 可配置的自动备份计划

- **增量备份**：
  - 仅备份自上次备份以来的新数据
  - 使用ID和时间戳追踪已备份内容
  - 减少数据传输和存储需求

### 3. 恢复功能

- **本地恢复**：
  - 从JSON文件恢复数据
  - 支持选择性恢复（仅短信、仅通话记录）
  - 预览恢复内容
  - ✅ 多组件状态同步，确保恢复过程一致性
  
- **远程恢复**：
  - 从服务器获取备份
  - 显示可用备份列表和时间
  - 支持跨设备恢复

- **恢复策略**：
  - 处理重复数据情况
  - 冲突解决机制
  - ✅ 恢复进度和结果报告
  - ✅ 备份文件预检查和内容分析
  - ✅ 组件间状态同步（MainActivity与RestoreViewModel）

### 4. 用户界面

- **Material Design 3实现**：
  - 遵循MD3设计语言和组件
  - 支持动态颜色（Android 12+）
  - 兼容性主题（旧版Android）
  - 自适应布局（手机和平板）

- **主界面**：
  - 清晰的备份/恢复操作入口
  - 上次备份状态显示
  - 数据统计概览
  
- **设置界面**：
  - 服务器配置
  - 备份策略设置
  - 语言选择
  - 主题定制

- **数据预览**：
  - 短信和通话记录列表视图
  - 搜索和过滤功能
  - 详情视图

### 5. 多语言支持

- **国际化框架**：
  - 字符串资源外部化
  - 内置维护中文（默认和 zh-rCN）与英文（en）
  - 用户可见文案必须进入 Android string resources，不在 Kotlin/Compose 代码中硬编码
  - 可扩展架构，便于社区后续添加更多语言
  
- **语言切换**：
  - 运行时切换界面语言
  - 保存用户语言偏好
  - 尊重系统语言设置

### 6. 安全与隐私

- **数据加密**：
  - 备份文件加密（AES-256）
  - 网络传输加密（HTTPS/TLS）
  - 可选的额外密码保护

- **隐私保护**：
  - 明确的权限请求和使用说明
  - 最小化权限要求
  - 无第三方服务依赖

- **认证与授权**：
  - 安全令牌管理
  - 可配置的服务器身份验证
  - 会话管理

### 7. 日志与监控

- **日志系统**：
  - 符合NOTICE.md规范的日志格式
  - 多级日志（ERROR、DEBUG、INFO）
  - 文件和控制台输出

- **错误处理**：
  - 优雅失败机制
  - 用户友好的错误消息
  - 错误追踪和上下文

- **性能监控**：
  - 关键操作计时
  - 内存使用监控
  - 电池消耗优化

## 项目架构

### 模块化架构

项目已从单模块架构重构为多模块架构，核心逻辑提取为独立的SDK模块，实现关注点分离和跨平台复用：

```
Commory Android/
├── sdk/backup/          # 纯Kotlin备份/恢复SDK
├── sdk/auth/            # 纯Kotlin认证组件
├── sdk/storage/         # Android Library存储组件
├── app/                 # Android应用壳
└── docs/architecture/   # 架构设计文档
```

#### SDK模块职责

| 模块 | 类型 | 职责 | 关键组件 |
|------|------|------|---------|
| `sdk/backup` | 纯Kotlin (java-library) | 备份/恢复核心逻辑 | `BackupManager`、`RestoreManager`、`BackupSerializer`、数据模型（`BackupData`、`Message`、`CallLog`、`Contact`）、读写接口（`SmsReader`/`SmsWriter`、`CallLogReader`/`CallLogWriter`、`ContactReader`/`ContactWriter`、`BackupFileReader`/`BackupFileWriter`） |
| `sdk/auth` | 纯Kotlin (java-library) | 认证与身份管理 | `AuthProvider`接口、`LocalAuthProvider`（本地模式）、`ThirdPartyAuthProvider`接口（第三方登录扩展）、`AuthManager`（认证管理器）、`UserInfo`、`AuthResult`、`AuthCredentials` |
| `sdk/storage` | Android Library | 存储抽象与实现 | `StorageProvider`接口、`LocalStorageProvider`（本地文件存储）、`RemoteStorageProvider`接口（远程存储）、Room数据库、Retrofit网络请求 |
| `app` | Android Application | 应用壳 | UI（Jetpack Compose）、导航、权限管理、Android平台实现（`AndroidSmsReader`/`AndroidSmsWriter`等）、ViewModel |

#### 模块依赖规则

- `app` → `sdk/backup`、`sdk/auth`、`sdk/storage`
- SDK模块之间**互不依赖**（`sdk/backup` 不依赖 `sdk/auth` 或 `sdk/storage`）
- SDK模块**不依赖** `app`
- `sdk/backup` 和 `sdk/auth` 为纯Kotlin模块，不包含任何Android依赖
- `sdk/storage` 为Android Library，依赖Room、Retrofit等Android库

#### 架构设计文档

详细的架构设计文档位于 `docs/architecture/` 目录：

- [鸿蒙适配设计文档](docs/architecture/harmonyos-adaptation.md) — 鸿蒙原生开发方案、KMP复用策略、权限差异分析
- [第三方登录设计文档](docs/architecture/third-party-auth.md) — AuthProvider接口设计、OAuth2.0扩展、Token安全
- [AI Agent集成设计文档](docs/architecture/ai-agent-integration.md) — 知识库架构、本地/云端AI模型、隐私保护
- [后端微服务架构设计文档](docs/architecture/backend-microservices.md) — 服务拆分、API网关、NAS自部署方案

## 技术实现原则

### 1. 架构设计

- **模块化**：
  - 清晰的关注点分离，核心逻辑与平台实现解耦
  - SDK模块采用接口驱动设计，平台实现通过依赖注入
  - 松耦合组件设计，SDK模块间零依赖
  - 纯Kotlin SDK模块可跨平台复用（Android、鸿蒙、JVM）

- **数据模型**：
  - SDK层数据模型：平台无关的纯数据类（`sdk/backup/model/`）
  - 应用层UI模型：专用于UI展示（`app/model/`、`app/models/`）
  - 存储层实体：本地数据库（`sdk/storage` Room实体）
  - 使用Kotlin数据类和扩展函数
  - 类型安全的转换逻辑
  - JSON序列化（使用Gson，封装在`BackupSerializer`中）
  - 版本化模型设计
  - 清晰的模型间映射关系

- **API设计**：
  - SDK接口定义平台无关契约（`SmsReader`、`StorageProvider`、`AuthProvider`）
  - RESTful接口
  - 使用Retrofit进行API调用
  - 可重试和缓存策略

### 2. UI框架

- **Jetpack Compose**：
  - 声明式UI设计
  - 组件重用
  - 状态管理
  - 动画和过渡

- **Material 3组件**：
  - 实现最新MD3组件库
  - 动态颜色支持
  - 可访问性合规
  - 黑暗模式支持

### 3. 性能考量

- **后台处理**：
  - 使用协程进行异步操作
  - WorkManager用于计划任务
  - 分页加载大型数据集

- **内存管理**：
  - 避免内存泄漏
  - 处理配置变更
  - 大数据集流式处理

- **电池优化**：
  - 高效网络请求
  - 优化后台处理
  - 低功耗设计原则

### 4. 测试策略

- **SDK模块测试**：
  - `sdk/backup`：纯Kotlin单元测试，无需Android模拟器，覆盖`BackupManager`、`RestoreManager`、`BackupSerializer`及数据模型
  - `sdk/auth`：纯Kotlin单元测试，覆盖`AuthProvider`接口契约、`LocalAuthProvider`逻辑、`AuthManager`委托模式
  - `sdk/storage`：Android Instrumented测试，验证`LocalStorageProvider`文件操作、Room数据库交互

- **应用层测试**：
  - 使用JUnit和Mockito
  - 模拟SDK接口进行ViewModel测试
  - 模拟ContentProvider访问
  - 业务逻辑全覆盖

- **集成测试**：
  - 使用AndroidX Test框架
  - Espresso UI测试
  - 端到端备份/恢复流程测试
  - SDK与App集成验证

- **CI/CD就绪**：
  - 测试可自动化
  - 关键流程回归测试

### 5. 可扩展性

- **SDK接口扩展**：
  - `SmsReader`/`SmsWriter`等接口可适配不同平台实现（Android、鸿蒙）
  - `AuthProvider`接口支持新增第三方登录提供商
  - `StorageProvider`接口支持新增存储后端（本地、云端、NAS）
  - `ThirdPartyAuthProvider`接口为OAuth2.0扩展预留

- **插件架构**：
  - 为未来功能预留扩展点
  - 可替换的服务实现
  - 接口而非具体类

- **设计模式**：
  - 适配器用于不同数据源
  - 策略模式用于可切换行为
  - 观察者用于事件通知
  - 委托模式用于`AuthManager`

- **未来扩展考虑**：
  - MMS消息支持
  - 联系人备份
  - 本地AI分析
  - 第三方服务连接选项
  - 鸿蒙原生适配（复用纯Kotlin SDK模块）

## 从类似应用的见解

1. **SQL vs. JSON格式**：
   - SQLite适合本地查询和索引
   - JSON适合服务器传输和跨平台兼容
   - 我们选择JSON作为主要格式，更适合互操作性和API传输

2. **批量处理与流处理**：
   - 批量读取适用于小型数据集
   - 流处理和分页适用于大型数据集
   - 我们实现混合方法以适应不同数据量

3. **UI/UX最佳实践**：
   - 明确操作流程和反馈
   - 进度指示器对长时间操作至关重要
   - 简洁明了的统计和摘要信息

4. **权限处理策略**：
   - 逐步请求权限优于一次性请求
   - 明确解释每个权限的用途
   - 处理权限被拒绝的优雅降级

5. **离线优先设计**：
   - 本地备份优先，远程备份可选
   - 断网时能正常工作的核心功能
   - 稳健的错误处理和重试机制

## 优先级列表

### ✅ 阶段0（SDK提取 — 已完成）：
1. ✅ 提取 `sdk/backup` — 纯Kotlin备份/恢复SDK（BackupManager、RestoreManager、BackupSerializer、数据模型、读写接口）
2. ✅ 提取 `sdk/auth` — 纯Kotlin认证组件（AuthProvider接口、LocalAuthProvider、ThirdPartyAuthProvider、AuthManager）
3. ✅ 提取 `sdk/storage` — Android Library存储组件（StorageProvider接口、LocalStorageProvider、RemoteStorageProvider、Room）
4. ✅ App模块适配 — Android平台实现注入SDK接口
5. ✅ 架构设计文档 — 鸿蒙适配、第三方登录、AI Agent、后端微服务

### 阶段1（核心功能）：
1. SMS和通话记录读取
2. 本地JSON备份
3. 基本Material Design 3界面
4. 权限处理和错误报告
5. 多语言支持（中文和英文）

### 阶段2（增强功能）：
1. 远程备份到服务器
2. 本地和远程恢复功能
3. 高级UI组件和数据预览
4. 增量备份策略
5. 备份加密

### 阶段3（高级功能）：
1. 自动备份计划
2. 数据分析和统计
3. 跨设备同步
4. 高级设置和自定义
5. 更多数据类型支持
