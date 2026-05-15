# Commory Android

Commory Android 是本地优先的客户端，负责采集、导出、恢复通信数据，并在用户选择时同步到 Commory Server。

## 身份与命名

- Application id：`com.iskenkenya.commory`
- App namespace：`com.iskenkenya.commory.mobile`
- SDK namespaces：`com.iskenkenya.commory.sdk.backup`、`com.iskenkenya.commory.sdk.auth`、`com.iskenkenya.commory.sdk.storage`
- 产品名：`Commory`

Go backend module name 与 Android package naming 相互独立，不与客户端 id 冲突。

## 运行模式

- `LOCAL_ONLY`：备份和恢复在设备本地运行，不需要 auth 或网络。
- `COMMORY_SERVER`：仍然先完成本地备份，然后在启用时执行已认证的服务器上传。

模式行为记录在 [../docs/android-runtime-modes.md](../docs/android-runtime-modes.md)。

## 模块

```text
android/
├── app/          # Compose UI、navigation、permissions、runtime mode、Android platform adapters
└── sdk/
    ├── backup/  # 纯 Kotlin 备份/恢复编排和 MsgLayer 映射
    ├── auth/    # 纯 Kotlin auth contracts
    └── storage/ # Android storage implementation 和 storage contracts
```

`app` 可以依赖 SDK 模块。SDK 模块不得依赖 `app`；纯 Kotlin SDK 模块不得 import Android framework APIs。

## 国际化

- 用户可见字符串必须放在 Android resource 中，不要在 Kotlin/Compose 代码里硬编码中文或英文。
- 内置语言只维护中文和英文：`values/strings.xml`、`values-en/strings.xml`、`values-zh-rCN/strings.xml`。
- Compose UI 使用 `stringResource`。
- 动态错误在 UI 边界前类型化，再用占位符本地化。

## 命令

```bash
./gradlew :app:compileDebugKotlin
./gradlew :app:testDebugUnitTest
./gradlew :sdk:backup:test
./gradlew :sdk:auth:test
```

从 `android/` 目录运行 Android checks。只有需要设备或 emulator 时才使用 instrumented tests。

## 文档

- 工程标准：[../docs/engineering-standards.md](../docs/engineering-standards.md)
- 移动端 API 契约：[../docs/mobile-api.md](../docs/mobile-api.md)
- 运行模式：[../docs/android-runtime-modes.md](../docs/android-runtime-modes.md)
- Changelog：[CHANGELOG.md](CHANGELOG.md)

项目不使用 `AI_EDIT_LOG.md` 这类历史报告。持久变更写入 `CHANGELOG.md`、PR 描述和 commits。
