# MessageVault Monorepo 迁移 Spec

## Why
MessageVault 组织下的 3 个仓库（.github、MessageVault-Mobile、SMS-Previewer）目前分散在独立仓库中，维护成本高、协作不便。将它们合并到当前工作区作为 Monorepo，推送到 git@github.com:IsKenKenYa/MessageVault.git，可以统一管理、简化 CI/CD 流程，并集中处理组织级文档和许可证。原始组织仓库保留不动，作为后续协作的预留。

## What Changes
- 将 `.github` 仓库中的 README.md、LICENSE 整合到 Monorepo 根目录
- 将 `MessageVault-Mobile` 仓库以保留 Git 历史的方式迁移至 `android/` 子目录
- 将 `SMS-Previewer` 仓库以保留 Git 历史的方式迁移至 `previewer/` 子目录
- 合并 `profile/README.md` 的内容到根 README.md
- 统一使用 GPL v3.0 许可证
- 在当前工作区 `/workspace` 初始化 Monorepo
- 初始化并编译各子项目（Android 项目、Vue 前端项目）
- 配置功能规划文档
- 将最终仓库推送到 git@github.com:IsKenKenYa/MessageVault.git
- 原始组织仓库（MessageVault/.github、MessageVault/MessageVault-Mobile、MessageVault/SMS-Previewer）保留不动，作为后续协作预留

## Impact
- Affected repos: MessageVault/.github, MessageVault/MessageVault-Mobile, MessageVault/SMS-Previewer
- Target: 当前工作区 `/workspace`，远程仓库 git@github.com:IsKenKenYa/MessageVault.git
- 所有源仓库的 Git 历史将被保留并合并到新仓库
- 许可证统一为 GPL v3.0
- 原始组织仓库不做任何修改，保留原状

## ADDED Requirements

### Requirement: Monorepo 初始化
系统 SHALL 在当前工作区 `/workspace` 初始化 Git 仓库作为 Monorepo 的基础。

#### Scenario: 成功初始化 Monorepo
- **WHEN** 执行 Monorepo 初始化流程
- **THEN** 当前工作区包含根目录文件（README.md、LICENSE、NOTICE.md）并完成初始提交

### Requirement: 仓库迁移保留 Git 历史
系统 SHALL 使用 `git filter-repo` 或 `git subtree` 等工具将每个源仓库迁移到指定子目录，同时保留完整的 Git 提交历史。

#### Scenario: 迁移 MessageVault-Mobile
- **WHEN** 将 MessageVault-Mobile 仓库迁移到 `android/` 子目录
- **THEN** `android/` 目录包含该仓库的所有文件，且 Git 历史完整保留

#### Scenario: 迁移 SMS-Previewer
- **WHEN** 将 SMS-Previewer 仓库迁移到 `previewer/` 子目录
- **THEN** `previewer/` 目录包含该仓库的所有文件，且 Git 历史完整保留

### Requirement: 组织配置整合
系统 SHALL 将 `.github` 仓库中的组织级文件整合到 Monorepo 根目录。

#### Scenario: README.md 合并
- **WHEN** 整合 `.github` 仓库的文档
- **THEN** 根 README.md 包含组织介绍（来自 profile/README.md）以及项目概览信息

#### Scenario: LICENSE 整合
- **WHEN** 整合许可证
- **THEN** 根目录和各子目录统一使用 GPL v3.0 许可证

### Requirement: 统一许可证
系统 SHALL 将所有项目统一使用 GPL v3.0 许可证，替换 SMS-Previewer 原有的 CC BY-NC-SA 4.0。

#### Scenario: 许可证统一
- **WHEN** Monorepo 构建完成
- **THEN** 根目录包含 GPL v3.0 LICENSE，`android/` 和 `previewer/` 子目录也使用 GPL v3.0

### Requirement: 项目初始化与编译
系统 SHALL 对各子项目进行依赖安装和编译验证。

#### Scenario: Android 项目编译
- **WHEN** 对 `android/` 子目录执行 Gradle 构建
- **THEN** 项目成功编译（或确认构建环境配置正确）

#### Scenario: Previewer 项目编译
- **WHEN** 对 `previewer/` 子目录执行 pnpm install && pnpm build
- **THEN** 项目成功编译

### Requirement: 功能规划配置
系统 SHALL 在 Monorepo 中配置功能规划文档，描述各子项目的功能定位和发展方向。

#### Scenario: 功能规划文档
- **WHEN** Monorepo 迁移完成
- **THEN** 根 README.md 或独立文档中包含各子项目的功能规划说明

### Requirement: 推送到远程仓库
系统 SHALL 将 Monorepo 推送到 git@github.com:IsKenKenYa/MessageVault.git。

> **注意**：在虚拟环境中不执行 git push 或分支切换操作，推送由用户手动完成。

#### Scenario: 推送成功
- **WHEN** Monorepo 迁移和验证完成
- **THEN** 代码已推送到 git@github.com:IsKenKenYa/MessageVault.git（由用户手动执行）

### Requirement: 原始仓库保留
系统 SHALL 不修改原始组织仓库（MessageVault/.github、MessageVault/MessageVault-Mobile、MessageVault/SMS-Previewer），仅以只读方式克隆并迁移内容。

#### Scenario: 原始仓库不受影响
- **WHEN** Monorepo 迁移流程执行
- **THEN** 原始组织仓库不做任何修改，保留原状作为后续协作预留

## MODIFIED Requirements

（无修改的需求）

## REMOVED Requirements

### Requirement: CC BY-NC-SA 4.0 许可证
**Reason**: 所有项目统一使用 GPL v3.0 许可证
**Migration**: 将 SMS-Previewer 的许可证从 CC BY-NC-SA 4.0 替换为 GPL v3.0
