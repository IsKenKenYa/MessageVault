# MessageVault Mobile

<div align="center">
  <img src="app/src/main/res/drawable/ic_launcher_foreground.xml" width="100" height="100" alt="MessageVault Logo">
</div>

<div align="center">
  <h2>安全备份和恢复Android短信和通话记录</h2>
</div>

<div align="center">
  <p>使用Material You设计的现代化备份解决方案</p>
</div>

<div align="center">

[![构建状态](https://img.shields.io/github/actions/workflow/status/MessageVault/MessageVault-Mobile/android.yml?branch=main)](https://github.com/MessageVault/MessageVault-Mobile/actions)
[![版本](https://img.shields.io/github/v/release/MessageVault/MessageVault-Mobile?include_prereleases)](https://github.com/MessageVault/MessageVault-Mobile/releases)
[![API](https://img.shields.io/badge/API-24%2B-brightgreen.svg)](https://android-arsenal.com/api?level=24)
[![许可证](https://img.shields.io/github/license/MessageVault/MessageVault-Mobile)](https://github.com/MessageVault/MessageVault-Mobile/blob/main/LICENSE)
[![Stars](https://img.shields.io/github/stars/MessageVault/MessageVault-Mobile)](https://github.com/MessageVault/MessageVault-Mobile/stargazers)
[![Issues](https://img.shields.io/github/issues/MessageVault/MessageVault-Mobile)](https://github.com/MessageVault/MessageVault-Mobile/issues)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/MessageVault/MessageVault-Mobile/pulls)

</div>

## ✨ 特性

- 📱 **现代化界面**: 采用Material You设计语言，支持动态取色
- 🔒 **安全可靠**: 本地优先，无需云存储，保护隐私
- 📤 **全面备份**: 支持短信、通话记录和联系人的备份
- 📥 **智能恢复**: 支持选择性恢复，带进度显示
- 🌍 **多语言**: 支持中文和英文界面
- 🔄 **状态同步**: 组件间状态实时同步，确保操作流畅

## 📱 功能概览

MessageVault Mobile是一个专注于Android数据备份与恢复的开源应用。它提供了一个现代化的解决方案，帮助用户安全地管理他们的短信、通话记录和联系人数据。

### 当前功能状态

| 功能 | 状态 | 描述 |
|------|------|------|
| 短信备份 | ✅ | 完全支持，包含完整元数据 |
| 通话记录备份 | ✅ | 完全支持，包含详细通话信息 |
| 联系人备份 | ✅ | 完全支持，包含所有联系人字段 |
| 短信恢复 | ✅ | 完全支持，带进度显示 |
| 通话记录恢复 | ✅ | 完全支持，带状态反馈 |
| 联系人恢复 | ⏳ | 开发中，基础框架已完成 |
| 权限管理 | ✅ | 完全支持，包含运行时权限处理 |

## 🛠️ 技术栈

- **UI框架**: Jetpack Compose
- **架构模式**: MVVM + 多模块SDK架构
- **状态管理**: ViewModel + StateFlow
- **依赖注入**: 构造函数注入（SDK接口由App模块实现注入）
- **数据持久化**: Room + DataStore
- **并发处理**: Kotlin Coroutines
- **序列化**: Gson（封装在BackupSerializer中）
- **网络请求**: Retrofit
- **单元测试**: JUnit + Mockito + Kotlin Coroutines Test

## 🏗️ 项目架构

### 模块结构

```
MessageVault-Mobile/
├── sdk/
│   ├── backup/                # 纯Kotlin备份/恢复SDK
│   │   └── src/main/kotlin/imken/messagevault/sdk/backup/
│   │       ├── BackupManager.kt       # 备份门面类
│   │       ├── RestoreManager.kt      # 恢复管理器
│   │       ├── model/                 # 数据模型（BackupData, Message, CallLog, Contact...）
│   │       ├── serializer/            # JSON序列化（BackupSerializer）
│   │       ├── reader/                # 读取接口（SmsReader, CallLogReader, ContactReader, BackupFileReader）
│   │       └── writer/                # 写入接口（SmsWriter, CallLogWriter, ContactWriter, BackupFileWriter）
│   ├── auth/                  # 纯Kotlin认证组件
│   │   └── src/main/kotlin/imken/messagevault/sdk/auth/
│   │       ├── AuthProvider.kt        # 认证接口 + 数据模型（UserInfo, AuthResult, AuthCredentials）
│   │       ├── LocalAuthProvider.kt   # 本地模式实现
│   │       ├── ThirdPartyAuthProvider.kt # 第三方登录接口
│   │       └── AuthManager.kt         # 认证管理器（委托模式）
│   └── storage/               # Android Library存储组件
│       └── src/main/kotlin/imken/messagevault/sdk/storage/
│           ├── StorageProvider.kt     # 存储接口 + 数据模型（StorageResult, StorageInfo）
│           ├── LocalStorageProvider.kt  # 本地文件存储实现
│           └── RemoteStorageProvider.kt # 远程存储接口
├── app/                       # Android应用壳
│   └── src/main/java/imken/messagevault/mobile/
│       ├── data/backup/       # Android平台SDK实现（AndroidSmsReader, AndroidCallLogWriter...）
│       ├── model/             # 应用层数据模型
│       ├── models/            # UI数据模型
│       ├── ui/                # Compose UI + ViewModel + 导航
│       ├── utils/             # 工具类
│       └── MainActivity.kt    # 主Activity
└── docs/architecture/         # 架构设计文档
    ├── harmonyos-adaptation.md
    ├── third-party-auth.md
    ├── ai-agent-integration.md
    └── backend-microservices.md
```

### 模块依赖关系

```
┌─────────────────────────────────────┐
│              app                     │
│  (UI · ViewModel · 权限 · 导航)      │
│  (Android平台SDK接口实现)             │
└──────┬──────────┬──────────┬────────┘
       │          │          │
       ▼          ▼          ▼
┌────────────┐ ┌────────────┐ ┌────────────┐
│ sdk/backup │ │  sdk/auth  │ │ sdk/storage │
│ (纯Kotlin) │ │ (纯Kotlin) │ │ (Android   │
│            │ │            │ │  Library)  │
└────────────┘ └────────────┘ └────────────┘
  备份/恢复      认证/登录      存储/数据库
  数据模型       身份管理       文件/远程存储
  序列化         Token管理      Room数据库
```

> SDK模块之间互不依赖，纯Kotlin模块（backup、auth）可跨平台复用

### 架构设计文档

| 文档 | 说明 |
|------|------|
| [鸿蒙适配设计文档](docs/architecture/harmonyos-adaptation.md) | 鸿蒙原生开发方案、KMP复用策略、权限差异分析 |
| [第三方登录设计文档](docs/architecture/third-party-auth.md) | AuthProvider接口设计、OAuth2.0扩展、Token安全 |
| [AI Agent集成设计文档](docs/architecture/ai-agent-integration.md) | 知识库架构、本地/云端AI模型、隐私保护 |
| [后端微服务架构设计文档](docs/architecture/backend-microservices.md) | 服务拆分、API网关、NAS自部署方案 |

## 📦 安装要求

- Android 7.0 (API 24) 或更高版本
- 约20MB存储空间
- 必要权限：
  - 读取/写入短信
  - 读取/写入通话记录
  - 读取/写入联系人
  - 存储访问（用于备份文件）

## 🚀 开始使用

### 开发环境配置

```bash
# 克隆仓库
git clone https://github.com/MessageVault/MessageVault-Mobile.git

# 进入项目目录
cd MessageVault-Mobile

# 构建整个项目（包含所有SDK模块和App）
./gradlew build

# 仅构建特定模块
./gradlew :sdk:backup:build     # 备份SDK
./gradlew :sdk:auth:build       # 认证组件
./gradlew :sdk:storage:build    # 存储组件
./gradlew :app:assembleDebug    # Debug APK

# 运行SDK模块测试（无需Android设备/模拟器）
./gradlew :sdk:backup:test      # 备份SDK单元测试
./gradlew :sdk:auth:test        # 认证组件单元测试

# 运行所有测试
./gradlew test
```

### 系统要求

- JDK 17 (推荐使用OpenJDK 17.0.14)
- Android Studio Iguana或更高版本
- Android SDK (API 24-34)
- Gradle 8.2+

### 模块说明

| 模块 | 构建类型 | 说明 | 测试方式 |
|------|---------|------|---------|
| `sdk/backup` | java-library | 纯Kotlin，无Android依赖 | `./gradlew :sdk:backup:test` |
| `sdk/auth` | java-library | 纯Kotlin，无Android依赖 | `./gradlew :sdk:auth:test` |
| `sdk/storage` | Android Library | 依赖Room、Retrofit等 | `./gradlew :sdk/storage:connectedAndroidTest` |
| `app` | Android Application | 依赖所有SDK模块 | `./gradlew :app:test` |

## 🤝 贡献指南

我们欢迎各种形式的贡献！以下是一些参与项目的方式：

- 🐛 提交bug报告
- 💡 提出新功能建议
- 📝 改进文档
- 🔍 审查代码

在提交Pull Request之前，请确保：

1. 遵循项目的代码风格
2. 添加必要的测试
3. 更新相关文档
4. 描述清楚改动的目的和影响

## 📄 开源协议

本项目采用[GNU General Public License v3.0](LICENSE)开源协议。

## 🙏 鸣谢

- [Jetpack Compose](https://developer.android.com/jetpack/compose)
- [Material Design 3](https://m3.material.io/)
- [Android Jetpack](https://developer.android.com/jetpack)

## 📬 联系我们

- 提交Issue: [GitHub Issues](https://github.com/MessageVault/MessageVault-Mobile/issues)
- 项目讨论: [GitHub Discussions](https://github.com/MessageVault/MessageVault-Mobile/discussions)