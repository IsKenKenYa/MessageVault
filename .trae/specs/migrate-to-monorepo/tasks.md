# Tasks

- [x] Task 1: 环境准备 — 安装必要工具，配置 Git 用户信息
  - [x] SubTask 1.1: 检查并安装 git-filter-repo（pip install git-filter-repo）
  - [x] SubTask 1.2: 配置 Git 用户名和邮箱
  - [x] SubTask 1.3: 检查 pnpm、Gradle 等构建工具是否可用

- [x] Task 2: 克隆所有源仓库到临时目录（只读克隆，不修改原始仓库）
  - [x] SubTask 2.1: 克隆 .github 仓库（main 分支）到 /tmp/msgvault-dotgithub
  - [x] SubTask 2.2: 克隆 MessageVault-Mobile 仓库（master 分支）到 /tmp/msgvault-mobile
  - [x] SubTask 2.3: 克隆 SMS-Previewer 仓库（MainMaster 分支）到 /tmp/msgvault-previewer

- [x] Task 3: 在当前工作区初始化 Monorepo 并整合 .github 仓库内容
  - [x] SubTask 3.1: 在 /workspace 执行 git init
  - [x] SubTask 3.2: 将 .github 仓库的根文件（LICENSE、NOTICE.md）复制到 /workspace 根目录
  - [x] SubTask 3.3: 合并 .github/profile/README.md 和 .github/README.md 的内容，生成根 README.md（含组织介绍 + 项目概览 + 许可证说明 + 功能规划）
  - [x] SubTask 3.4: 确保根目录 LICENSE 为 GPL v3.0
  - [x] SubTask 3.5: 提交初始根目录文件

- [x] Task 4: 迁移 MessageVault-Mobile 到 android/ 子目录（保留历史）
  - [x] SubTask 4.1: 使用 git filter-repo 将 MessageVault-Mobile 的文件重写到 android/ 前缀下
  - [x] SubTask 4.2: 将重写后的仓库作为 remote 添加到 Monorepo 并 merge
  - [x] SubTask 4.3: 确保 android/ 目录下使用 GPL v3.0 LICENSE

- [x] Task 5: 迁移 SMS-Previewer 到 previewer/ 子目录（保留历史）
  - [x] SubTask 5.1: 使用 git filter-repo 将 SMS-Previewer 的文件重写到 previewer/ 前缀下
  - [x] SubTask 5.2: 将重写后的仓库作为 remote 添加到 Monorepo 并 merge
  - [x] SubTask 5.3: 将 previewer/ 的许可证从 CC BY-NC-SA 4.0 替换为 GPL v3.0

- [x] Task 6: 初始化并编译各子项目
  - [x] SubTask 6.1: 在 previewer/ 目录执行 pnpm install && pnpm build
  - [x] SubTask 6.2: 在 android/ 目录执行 ./gradlew build（或确认构建环境配置正确）

- [x] Task 7: 配置功能规划
  - [x] SubTask 7.1: 在根 README.md 中添加各子项目的功能定位和发展方向说明

- [x] Task 8: 验证迁移结果
  - [x] SubTask 8.1: 验证目录结构正确（android/、previewer/、根文件）
  - [x] SubTask 8.2: 验证 Git 历史包含所有源仓库的提交
  - [x] SubTask 8.3: 验证所有目录统一使用 GPL v3.0 许可证
  - [x] SubTask 8.4: 验证根 README.md 内容完整（含功能规划）
  - [x] SubTask 8.5: 验证各子项目编译成功

- [ ] Task 9: 推送到远程仓库（由用户手动执行，虚拟环境中不执行 git push）
  - [ ] SubTask 9.1: 添加远程仓库 git@github.com:IsKenKenYa/MessageVault.git
  - [ ] SubTask 9.2: 推送代码到远程仓库

# Task Dependencies
- [Task 2] depends on [Task 1]
- [Task 3] depends on [Task 2]
- [Task 4] depends on [Task 3]
- [Task 5] depends on [Task 3]（Task 4 和 Task 5 可并行）
- [Task 6] depends on [Task 4, Task 5]
- [Task 7] depends on [Task 3]
- [Task 8] depends on [Task 6, Task 7]
- [Task 9] depends on [Task 8]
