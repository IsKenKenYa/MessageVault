# MsgLayer 索引说明

`MsgLayer` 把 canonical exported data 与查询时索引分开。

## Canonical Source

- root export JSON
- identities
- events

## 首批查询索引

- 跨 SMS text、voice transcript、voice summary 的 keyword search
- participant/contact grouping
- 按 event timestamp 排序的 timeline
- 从 `same_thread` 重建 thread

## 存储方向

首个 backend 版本支持：

- SQLite：用于 local-first deployment
- PostgreSQL：用于 service-style deployment

schema 已做基础规范化，以便未来增加 adapter 时不改变导出的 `MsgLayer` wire format。
