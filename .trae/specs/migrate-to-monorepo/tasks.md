# Tasks

- [ ] Task 1: 环境准备 — 安装 git-filter-repo 工具，配置 Git 用户信息
  - [ ] SubTask 1.1: 检查并安装 git-filter-repo（pip install git-filter-repo）
  - [ ] SubTask 1.2: 配置 Git 用户名和邮箱（用于提交）

- [ ] Task 2: 克隆所有源仓库到临时目录
  - [ ] SubTask 2.1: 克隆 .github 仓库（main 分支）
  - [ ] SubTask 2.2: 克隆 MessageVault-Mobile 仓库（master 分支）
  - [ ] SubTask 2.3: 克隆 SMS-Previewer 仓库（MainMaster 分支）

- [ ] Task 3: 初始化 Monorepo 并整合 .github 仓库内容
  - [ ] SubTask 3.1: 创建新的 Git 仓库（git init）
  - [ ] SubTask 3.2: 将 .github 仓库的根文件（LICENSE、NOTICE.md）复制到 Monorepo 根目录
  - [ ] SubTask 3.3: 合并 .github/profile/README.md 和 .github/README.md 的内容，生成根 README.md（含组织介绍 + 项目概览 + 许可证说明）
  - [ ] SubTask 3.4: 提交初始根目录文件

- [ ] Task 4: 迁移 MessageVault-Mobile 到 mobile/ 子目录（保留历史）
  - [ ] SubTask 4.1: 使用 git filter-repo 将 MessageVault-Mobile 的文件重写到 mobile/ 前缀下
  - [ ] SubTask 4.2: 将重写后的仓库作为 remote 添加到 Monorepo 并 merge
  - [ ] SubTask 4.3: 确保 mobile/ 目录下包含 GPL v3.0 LICENSE 文件

- [ ] Task 5: 迁移 SMS-Previewer 到 previewer/ 子目录（保留历史）
  - [ ] SubTask 5.1: 使用 git filter-repo 将 SMS-Previewer 的文件重写到 previewer/ 前缀下
  - [ ] SubTask 5.2: 将重写后的仓库作为 remote 添加到 Monorepo 并 merge
  - [ ] SubTask 5.3: 确保 previewer/ 目录下包含 CC BY-NC-SA 4.0 LICENSE 文件

- [ ] Task 6: 推送到目标远程仓库
  - [ ] SubTask 6.1: 添加远程仓库 git@github.com:IsKenKenYa/MessageVault.git
  - [ ] SubTask 6.2: 推送所有分支和历史到远程仓库

- [ ] Task 7: 验证迁移结果
  - [ ] SubTask 7.1: 验证目录结构正确（mobile/、previewer/、根文件）
  - [ ] SubTask 7.2: 验证 Git 历史包含所有源仓库的提交
  - [ ] SubTask 7.3: 验证各子目录的许可证文件存在
  - [ ] SubTask 7.4: 验证根 README.md 内容完整

# Task Dependencies
- [Task 2] depends on [Task 1]
- [Task 3] depends on [Task 2]
- [Task 4] depends on [Task 3]
- [Task 5] depends on [Task 3]（Task 4 和 Task 5 可并行）
- [Task 6] depends on [Task 4, Task 5]
- [Task 7] depends on [Task 6]
