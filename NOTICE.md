# Commory 治理声明

Commory 是一个本地优先的通信记忆系统。项目把短信、通话记录、联系人和未来可扩展的通信来源转换为结构化 MsgLayer 数据，用户可以保存在本地，也可以选择同步到自托管服务器。

## 原则

- 本地优先：Android 在没有服务器账号或网络时也必须可用。
- 服务器可选：Commory Server 提供同步、搜索和远程恢复能力，但不能成为本地备份的隐藏前置条件。
- 隐私可控：日志、错误、分析、文档和测试不得暴露短信正文、联系人内容、access token、refresh token 或私有备份载荷。
- 契约驱动：Android、backend、web、MsgLayer 的跨边界变更必须同步更新文档和测试。
- 中文优先但不放弃国际化：维护文档和注释优先中文；产品用户可见文案必须走中英文 i18n/resource。

## 仓库边界

- `android/`：Commory Android 客户端和 SDK 模块。
- `backend/`：自托管 Commory Server。
- `web/`：server-backed workflow 使用的 Vue dashboard。
- `msglayer/`：canonical interchange schema 和 examples。
- `docs/`：当前工程和 API 文档。
- `previewer/`：历史 XML SMS viewer 归档；当前工作不要更新。
- `references/`：只读外部参考代码；不要在那里开发功能。

## 变更历史

使用标准 changelog 和 review 渠道：

- 持久发布历史写入对应的 `CHANGELOG.md`。
- Pull request 和 commit 说明实现细节、测试情况，以及必要的 AI 协助说明。
- 不要创建或更新 `AI_EDIT_LOG.md`；AI 协助不使用专门旁路日志记录。

## 标准

`docs/engineering-standards.md` 是唯一工程标准。它覆盖命名、包结构、skills、测试、日志、上下文管理、国际化边界和反膨胀规则。

## 许可证

仓库根目录采用 GPL v3。`references/` 下的外部代码保持其上游许可证，并不会自动成为 Commory 可分发源码的一部分。
