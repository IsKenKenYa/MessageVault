# MsgLayer 版本管理

## 当前版本

- `msglayer/v0.1`

## 规则

- 向后兼容的 additive fields 保持在同一 major/minor version line 内。
- 破坏性结构变更必须使用新的 schema version。
- Schema files、examples、Android mappers 和 Go types 必须一起演进。
- `indexes` 是可选派生数据，绝不能成为 canonical source of truth。

## 兼容性意图

`v0.1` 是 Commory 的第一个稳定 interchange format。

- Android export 切换到 `MsgLayer`
- Go backend import/validation 目标为 `MsgLayer`
- legacy backup JSON 只通过 bridge converters 保持 restore-compatible

## v0.1 Relation Type 状态

- 当前 code paths 已启用：`same_thread`、`references_identity`
- 为未来 producers/consumers 预留：`reply_to`、`derived_from`
