# 变更描述

请说明本 PR 解决的问题、用户可见变化和主要实现方式。

## 影响范围

- [ ] Android
- [ ] Backend
- [ ] Web
- [ ] MsgLayer/schema
- [ ] Agent/runtime
- [ ] Docker/部署
- [ ] 文档/Rules

## 契约一致性

- [ ] API 变更已同步 `docs/mobile-api.md`
- [ ] MsgLayer 变更已同步 schema、examples、Android mapper 和 Go types
- [ ] Web 用户可见文案已同步 `zh.json` 和 `en.json`
- [ ] Android 用户可见文案已同步 `values/`、`values-en/`、`values-zh-rCN/`

## 隐私与本地模式

- [ ] Local-only mode 不需要网络或 auth
- [ ] 日志不包含短信正文、联系人内容、token、备份载荷或带凭据 URL
- [ ] Agent context policy 符合最小必要上下文原则

## 测试

请列出已运行的命令和结果：

```bash

```

## 发布与兼容性

- [ ] 需要更新 `CHANGELOG.md`
- [ ] 需要迁移说明
- [ ] 需要 Docker Compose 验证
- [ ] 跨 Android + Backend + Web + MsgLayer 的变更已完成多端检查
