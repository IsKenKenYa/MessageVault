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
- Android: MVVM + Repository Pattern
- Vue: Composition API + 单向数据流
- 模块化设计，职责分离

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
- 测试覆盖率目标：核心模块 > 70%

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
├── android/          # MessageVault-Mobile (Kotlin/Android)
├── previewer/        # SMS-Previewer (Vue 3/Vite)
├── .trae/            # 开发工具配置
├── LICENSE           # GPL v3.0
├── NOTICE.md         # 项目规范
└── README.md         # 项目总览
```
