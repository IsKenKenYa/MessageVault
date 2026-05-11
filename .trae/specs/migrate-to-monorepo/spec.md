# MessageVault Monorepo 迁移 Spec

## Why
MessageVault 组织下的 3 个仓库（.github、MessageVault-Mobile、SMS-Previewer）目前分散在独立仓库中，维护成本高、协作不便。将它们合并为一个 Monorepo 可以统一管理、简化 CI/CD 流程，并集中处理组织级文档和许可证。

## What Changes
- 将 `.github` 仓库中的 README.md、LICENSE 整合到新 Monorepo 根目录
- 将 `MessageVault-Mobile` 仓库以保留 Git 历史的方式迁移至 `mobile/` 子目录
- 将 `SMS-Previewer` 仓库以保留 Git 历史的方式迁移至 `previewer/` 子目录
- 合并 `profile/README.md` 的内容到根 README.md
- 处理多许可证问题（GPL v3.0 与 CC BY-NC-SA 4.0 共存）
- 推送最终结果到 `git@github.com:IsKenKenYa/MessageVault.git`

## Impact
- Affected repos: MessageVault/.github, MessageVault/MessageVault-Mobile, MessageVault/SMS-Previewer
- Target repo: IsKenKenYa/MessageVault
- 所有源仓库的 Git 历史将被保留并合并到新仓库
- 许可证文件需要在各子目录中保留各自的 LICENSE

## ADDED Requirements

### Requirement: Monorepo 初始化
系统 SHALL 创建一个新的 Git 仓库作为 Monorepo 的基础，并推送到 `git@github.com:IsKenKenYa/MessageVault.git`。

#### Scenario: 成功初始化 Monorepo
- **WHEN** 执行 Monorepo 初始化流程
- **THEN** 新仓库包含根目录文件（README.md、LICENSE、NOTICE.md）并推送到远程仓库

### Requirement: 仓库迁移保留 Git 历史
系统 SHALL 使用 `git filter-repo` 或 `git subtree` 等工具将每个源仓库迁移到指定子目录，同时保留完整的 Git 提交历史。

#### Scenario: 迁移 MessageVault-Mobile
- **WHEN** 将 MessageVault-Mobile 仓库迁移到 `mobile/` 子目录
- **THEN** `mobile/` 目录包含该仓库的所有文件，且 Git 历史完整保留

#### Scenario: 迁移 SMS-Previewer
- **WHEN** 将 SMS-Previewer 仓库迁移到 `previewer/` 子目录
- **THEN** `previewer/` 目录包含该仓库的所有文件，且 Git 历史完整保留

### Requirement: 组织配置整合
系统 SHALL 将 `.github` 仓库中的组织级文件整合到 Monorepo 根目录。

#### Scenario: README.md 合并
- **WHEN** 整合 `.github` 仓库的文档
- **THEN** 根 README.md 包含组织介绍（来自 profile/README.md）以及项目概览信息

#### Scenario: LICENSE 整合
- **WHEN** 整合 `.github` 仓库的许可证
- **THEN** 根目录包含 LICENSE 文件（来自 .github 仓库）

### Requirement: 多许可证处理
系统 SHALL 在各子目录中保留各自的许可证文件，并在根 README.md 中说明许可证差异。

#### Scenario: 许可证共存
- **WHEN** Monorepo 包含不同许可证的项目
- **THEN** `mobile/` 目录包含 GPL v3.0 许可证，`previewer/` 目录包含 CC BY-NC-SA 4.0 许可证，根 README.md 中有许可证说明

### Requirement: 目标仓库推送
系统 SHALL 将最终的 Monorepo 推送到指定的远程仓库。

#### Scenario: 成功推送
- **WHEN** Monorepo 构建完成
- **THEN** 所有代码和历史已推送到 `git@github.com:IsKenKenYa/MessageVault.git` 的 main 分支

## MODIFIED Requirements

（无修改的需求）

## REMOVED Requirements

（无移除的需求）
