# Changelog

Commory 使用 `v0.x.y` 语义化版本。`v0` 阶段 API 和 schema 仍可能变化；每个版本必须记录用户可见变更、兼容性影响和迁移说明。

## Unreleased

### Added

- 单端口部署规划：Go backend 可同时提供 `/api` 和 Web 静态资源。
- 根目录 Docker Compose 开发部署入口。
- 项目管理、技术债和 Web dashboard 规范文档。

### Changed

- Web 生产环境 API 默认改为同源 `/`，避免生产部署中的 CORS 配置负担。

### Compatibility

- 当前 Docker Compose 仍使用 file-backed sqlite 开发存储；真实 SQL 存储迁移以后再扩展 Postgres/MySQL 部署方案。
