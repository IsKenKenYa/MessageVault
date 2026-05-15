# Android -> MsgLayer v0.1 映射

本文定义当前 Android backup models 到 `msglayer/v0.1` 的第一版稳定映射。

## Root

- `version` -> 固定为 `msglayer/v0.1`
- `exported_at` -> RFC3339 UTC 格式的导出时间
- `source.platform` -> `android`
- `source.device_id` -> app layer 传入的 Android device identifier
- `source.app_version` -> app layer 传入的 Android app version

## Self Identity

每次 export 都创建一个稳定的 self identity：

- `id` -> `self/<device_id>`
- `type` -> `device`
- `display_name` -> 可用时使用 device info string
- `labels` -> `["self"]`

## Contact -> Identity

Android `Contact` 映射为 MsgLayer `identity`：

- `type` -> `person`
- `phones` -> `phoneNumbers`
- `emails` -> `emails`
- `avatar` -> `photoData`
- `meta.source` -> `contacts`

稳定 identity IDs 优先由规范化手机号生成；没有手机号时，退回使用 contact ID/name fingerprints。

## Contact -> contact_snapshot Event

contacts 导出的每个 identity 也会生成一个 `contact_snapshot` event：

- `direction` -> `system`
- `participants` -> contact identity itself
- `content.identity_id` -> identity ID
- `content.snapshot` -> identity payload
- `relations` -> `references_identity`

## Message -> sms Event

- `Message.id` -> event ID suffix
- `Message.date` -> event timestamp
- `Message.type` -> `direction`
  - `1` => `inbound`
  - `2` => `outbound`
  - `3/4/5/6` => 为了 schema compatibility 导出为 `outbound`，同时在 `meta.raw_type` 保留原始 Android type
  - fallback => `inbound`
- `Message.address` -> participant identity resolution input
- `Message.body` -> `content.text`
- `Message.threadId` -> `relations.same_thread`
- `Message.readState` -> `meta.read`
- `Message.messageStatus` -> `meta.status`

## CallLog -> call Event

- `CallLog.id` -> event ID suffix
- `CallLog.date` -> event timestamp
- `CallLog.number` / `CallLog.contact` -> participant identity resolution input
- `CallLog.type` -> `direction` 和 `content.call_type`
- `CallLog.duration` -> `content.duration_sec`
- `CallLog.contact` -> `meta.contact_name`

export 中的 timestamp ordering 依赖 mapper：它必须先把所有 event timestamps 规范化为 UTC RFC3339，再排序。

## Restore Boundary

`MsgLayer` 现在是主要 export format。

当前 restore compatibility 通过把 `MsgLayer` events 转换回现有 Android restore DTOs 来保持：

- `sms` -> `SmsData`
- `call` -> `CallLogData`
- `identity/contact_snapshot` -> `ContactData`

这样可以在 write path 迁移到新标准时，继续保持 restore 行为可用。

## Round-Trip 中已知数据损失

`MsgLayer -> legacy restore DTO` bridge 在 v0.1 中刻意优先兼容性，而不是完美保真。

当前已知损失：

- contact `groups`、`websites` 和 `note` 只保留在 identity metadata 中
- avatar/photo data 不会恢复进 legacy contact DTO
- legacy Android raw message/call types 保留为 metadata，而不是 first-class restore fields
- 非数字 event IDs 在 restore bridge 中会退回为派生 numeric IDs
