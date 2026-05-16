# 技术债清单

本文记录已经识别但不应混入当前功能 PR 的技术债。每项技术债必须绑定解决阶段或触发条件。

## Android Gradle 弃用警告

- 问题：Jetifier、Manifest package、Gradle 10 deprecation 等警告会影响后续构建升级。
- 影响：短期不阻塞功能，但会增加 Android toolchain 升级成本。
- 计划：独立 issue 跟踪，在下一轮 Android 构建系统整理中处理。

## `persist()` 全量写入

- 问题：当前 file-backed 存储在持久化时存在全量写入性能风险。
- 影响：数据量增长后导入、refresh token、查询写入路径可能变慢。
- 计划：ADR 0002 的 SQL 存储迁移阶段解决，通过真实 SQL migrations 和 typed query 替换。

## `ConsumeRefreshToken` 线性扫描

- 问题：当前 file-backed 存储中 refresh token 消费使用线性扫描。
- 影响：当前开发规模可接受，但多用户或长期运行后不适合生产。
- 计划：SQL 迁移时建立 refresh token 索引，并用 storage contract tests 固化行为。
