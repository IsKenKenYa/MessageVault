# 后端微服务架构设计文档

## 1. 概述

### 1.1 为什么需要后端微服务架构
Commory 目前以本地备份为核心，所有数据存储在设备端。以下需求驱动了后端微服务架构的引入：

| 需求 | 说明 |
|------|------|
| 云端同步 | 用户期望多设备间同步备份数据 |
| 跨设备恢复 | 换机时从云端恢复历史备份 |
| AI 云端分析 | 本地模型能力有限，需云端大模型支持 |
| 协作备份 | 家庭成员共享备份空间 |
| 数据安全 | 云端加密存储，防止设备丢失导致数据丢失 |

### 1.2 设计目标

- **轻量可自部署**：支持家庭 NAS 的 Docker Compose 一键部署
- **渐进式架构**：从单体开始，按需拆分为微服务
- **安全优先**：端到端加密、零知识架构、最小权限
- **成本可控**：资源占用低，适合个人/家庭使用场景

---

## 2. 服务拆分方案

### 2.1 服务全景图

```mermaid
graph TB
    subgraph client[客户端]
        android[Android]
        harmony[鸿蒙]
        web[Web 管理面板]
    end

    subgraph gateway[API 网关]
        gw[API Gateway<br/>Nginx / Kong]
    end

    subgraph services[微服务]
        auth[认证服务<br/>auth-service]
        backup[备份存储服务<br/>backup-service]
        ai[AI 分析服务<br/>ai-service]
        push[推送服务<br/>push-service]
    end

    subgraph storage[存储层]
        pg[(PostgreSQL)]
        minio[(MinIO)]
        redis[(Redis)]
        chroma[(ChromaDB)]
    end

    client --> gateway
    gateway --> auth & backup & ai & push
    auth --> pg & redis
    backup --> pg & minio
    ai --> pg & chroma
    push --> redis
```

### 2.2 服务职责

| 服务 | 职责 | 技术栈 | 数据存储 |
|------|------|--------|---------|
| auth-service | 用户认证、Token 管理、OAuth2.0 | Go / Kotlin | PostgreSQL + Redis |
| backup-service | 备份上传/下载/管理、版本控制 | Go / Kotlin | PostgreSQL + MinIO |
| ai-service | AI 分析、知识库管理、模型调度 | Python | PostgreSQL + ChromaDB |
| push-service | 消息推送、事件通知 | Go | Redis |
| API Gateway | 路由、限流、鉴权、日志 | Nginx / Kong | - |

### 2.3 服务间通信

```mermaid
graph LR
    gw[API Gateway] -->|HTTP/REST| auth & backup & ai & push
    auth -->|gRPC| backup
    backup -->|异步消息| push
    ai -->|HTTP/REST| backup
    subgraph mq[消息队列 - Redis Streams]
        stream1[backup.events]
        stream2[auth.events]
    end
    backup --> stream1
    auth --> stream2
    push --> stream1
```

### 2.4 服务 API 设计

#### auth-service

| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 |
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/auth/logout` | 用户登出 |
| POST | `/api/v1/auth/refresh` | 刷新 Token |
| POST | `/api/v1/auth/oauth2/{provider}` | OAuth2.0 登录 |
| GET | `/api/v1/auth/me` | 获取当前用户 |

#### backup-service

| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/api/v1/backups` | 上传备份 |
| GET | `/api/v1/backups` | 列出备份 |
| GET | `/api/v1/backups/{id}` | 获取备份详情 |
| GET | `/api/v1/backups/{id}/download` | 下载备份 |
| DELETE | `/api/v1/backups/{id}` | 删除备份 |

#### ai-service

| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/api/v1/ai/analyze` | 执行分析 |
| POST | `/api/v1/ai/analyze/stream` | 流式分析 |
| POST | `/api/v1/ai/knowledge-base/rebuild` | 重建知识库 |
| GET | `/api/v1/ai/models` | 获取可用模型 |

---

## 3. API 网关设计

### 3.1 网关架构

```mermaid
graph TB
    client[客户端请求] --> gateway[API Gateway - Kong]
    subgraph plugins[插件链]
        plugin_auth[JWT 认证]
        plugin_rate[限流控制]
        plugin_cors[CORS]
        plugin_log[请求日志]
    end
    gateway --> plugins --> router[路由引擎]
    router --> auth_svc[auth-service :8081] & backup_svc[backup-service :8082] & ai_svc[ai-service :8083] & push_svc[push-service :8084]
```

### 3.2 路由规则

| 路径 | 目标服务 | 认证要求 | 限流 |
|------|---------|---------|------|
| `/api/v1/auth/login` | auth-service | 无 | 10次/分钟 |
| `/api/v1/auth/register` | auth-service | 无 | 5次/分钟 |
| `/api/v1/auth/*` | auth-service | JWT | 60次/分钟 |
| `/api/v1/backups` | backup-service | JWT | 30次/分钟 |
| `/api/v1/backups/*/download` | backup-service | JWT | 10次/分钟 |
| `/api/v1/ai/*` | ai-service | JWT | 20次/分钟 |

### 3.3 鉴权流程

```mermaid
sequenceDiagram
    participant Client as 客户端
    participant GW as API Gateway
    participant Auth as auth-service
    participant Service as 目标服务

    Client->>GW: 请求 (携带 JWT)
    GW->>GW: 验证 JWT 签名和有效期
    alt JWT 有效
        GW->>Service: 转发请求 (X-User-Id Header)
        Service-->>Client: 200 OK
    else JWT 过期
        GW-->>Client: 401 Token Expired
        Client->>GW: POST /auth/refresh
        GW->>Auth: 验证 Refresh Token
        Auth-->>Client: 新 JWT
    else JWT 无效
        GW-->>Client: 401 Unauthorized
    end
```

---

## 4. 数据流架构

### 4.1 备份上传数据流

```mermaid
sequenceDiagram
    participant App as 移动端
    participant GW as API Gateway
    participant Backup as backup-service
    participant MinIO as MinIO
    participant PG as PostgreSQL

    App->>GW: POST /backups (JWT + 加密数据)
    GW->>Backup: 转发请求
    Backup->>Backup: 验证 checksum
    Backup->>MinIO: 上传加密文件
    MinIO-->>Backup: object_key
    Backup->>PG: 保存元数据
    Backup-->>App: 201 Created { backup_id }
```

### 4.2 AI 分析数据流

```mermaid
sequenceDiagram
    participant App as 移动端
    participant AI as ai-service
    participant Chroma as ChromaDB
    participant LLM as 云端 LLM

    App->>AI: POST /ai/analyze (JWT + 查询)
    AI->>Chroma: 语义检索
    Chroma-->>AI: 相关文档片段
    AI->>AI: 数据脱敏
    AI->>LLM: 调用大模型
    LLM-->>AI: 分析结果
    AI->>AI: 脱敏还原
    AI-->>App: 200 OK { analysis }
```

### 4.3 端到端加密

```mermaid
graph LR
    subgraph mobile[移动端]
        encrypt[AES-256-GCM 加密]
    end
    subgraph transport[传输层]
        https[HTTPS / TLS 1.3]
    end
    subgraph server[服务端]
        backup_svc[backup-service]
        minio[MinIO 加密存储]
    end
    encrypt --> https --> backup_svc --> minio
```

服务端无法解密用户数据，实现零知识架构。

---

## 5. 部署架构

### 5.1 Kubernetes 部署

```mermaid
graph TB
    subgraph k8s[Kubernetes 集群]
        ingress[Nginx Ingress]
        subgraph svc[服务层]
            auth_d[auth-service 2 replicas]
            backup_d[backup-service 2 replicas]
            ai_d[ai-service 1 replica]
            push_d[push-service 1 replica]
        end
        subgraph mw[中间件]
            pg_s[PostgreSQL]
            redis_s[Redis]
            minio_s[MinIO]
            chroma_s[ChromaDB]
        end
    end
    ingress --> auth_d & backup_d & ai_d & push_d
    auth_d --> pg_s & redis_s
    backup_d --> pg_s & minio_s
    ai_d --> pg_s & chroma_s
```
资源配置：auth-service (2副本, 100-500m CPU, 128-256Mi)、backup-service (2副本, 200-1000m, 256-512Mi)、ai-service (1副本, 500-2000m, 512Mi-2Gi)、push-service (1副本, 50-200m, 64-128Mi)。

---

## 6. NAS 自部署方案

### 6.1 Docker Compose 配置

```yaml
version: '3.8'
services:
  nginx:
    image: nginx:alpine
    ports: ["8443:443", "8080:80"]
    volumes: [./config/nginx/nginx.conf:/etc/nginx/nginx.conf:ro, ./config/nginx/ssl:/etc/nginx/ssl:ro]
    depends_on: [auth-service, backup-service, ai-service]
    restart: unless-stopped
  auth-service:
    image: commory/auth-service:latest
    environment: [DATABASE_URL=postgresql://mvault:mvault@postgres:5432/mvault_auth, REDIS_URL=redis://redis:6379/0, JWT_SECRET=${JWT_SECRET}]
    depends_on: [postgres, redis]
    restart: unless-stopped
  backup-service:
    image: commory/backup-service:latest
    environment: [DATABASE_URL=postgresql://mvault:mvault@postgres:5432/mvault_backup, MINIO_ENDPOINT=minio:9000, MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY}, MINIO_SECRET_KEY=${MINIO_SECRET_KEY}]
    depends_on: [postgres, minio]
    restart: unless-stopped
  ai-service:
    image: commory/ai-service:latest
    environment: [DATABASE_URL=postgresql://mvault:mvault@postgres:5432/mvault_ai, CHROMA_HOST=chromadb, CHROMA_PORT=8000, OPENAI_API_KEY=${OPENAI_API_KEY:-}]
    depends_on: [postgres, chromadb]
    restart: unless-stopped
  push-service:
    image: commory/push-service:latest
    environment: [REDIS_URL=redis://redis:6379/1]
    depends_on: [redis]
    restart: unless-stopped
  postgres:
    image: postgres:16-alpine
    environment: [POSTGRES_USER=mvault, POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-mvault}]
    volumes: [./data/postgres:/var/lib/postgresql/data]
    restart: unless-stopped
  redis:
    image: redis:7-alpine
    command: redis-server --maxmemory 128mb --maxmemory-policy allkeys-lru
    volumes: [./data/redis:/data]
    restart: unless-stopped
  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment: [MINIO_ROOT_USER=${MINIO_ACCESS_KEY:-minioadmin}, MINIO_ROOT_PASSWORD=${MINIO_SECRET_KEY:-minioadmin}]
    volumes: [./data/minio:/data]
    ports: ["9001:9001"]
    restart: unless-stopped
  chromadb:
    image: chromadb/chroma:latest
    environment: [ANONYMIZED_TELEMETRY=FALSE]
    volumes: [./data/chroma:/chroma/chroma]
    restart: unless-stopped
```

### 6.2 轻量部署模式

对于资源受限的 NAS（< 4GB RAM），提供轻量模式：合并 auth + backup + push 为单服务，使用 SQLite 替代 PostgreSQL，省略 ChromaDB（AI 仅支持云端模型）。

| 部署模式 | 内存占用 | 适用场景 |
|---------|---------|---------|
| 完整模式 | ~3GB | 高性能 NAS / 服务器 |
| 轻量模式 | ~1GB | 入门 NAS / 树莓派 |
| All-in-One | ~512MB | 极低资源环境 |

---

## 7. 开发路线图

### 7.1 阶段规划

```mermaid
gantt
    title 后端微服务开发路线图
    dateFormat YYYY-MM-DD
    section 阶段一：基础服务
    auth-service 开发           :a1, 2026-06-01, 30d
    backup-service 开发         :a2, 2026-06-15, 30d
    API Gateway 搭建            :a3, 2026-07-01, 15d
    section 阶段二：存储与同步
    MinIO 对象存储集成          :b1, 2026-07-15, 15d
    端到端加密实现              :b2, 2026-08-01, 20d
    section 阶段三：AI 服务
    ai-service 基础框架         :c1, 2026-08-15, 20d
    知识库与云端模型            :c2, 2026-09-01, 25d
    section 阶段四：部署运维
    Docker Compose 配置         :d1, 2026-10-01, 10d
    NAS 部署方案验证            :d2, 2026-10-10, 15d
    K8s Helm Chart             :d3, 2026-10-25, 15d
```

### 7.2 里程碑

| 里程碑 | 时间 | 交付物 |
|--------|------|--------|
| M1 - 认证就绪 | 2026-07 | auth-service + JWT + OAuth2.0 |
| M2 - 备份可用 | 2026-08 | backup-service + MinIO + 加密 |
| M3 - AI 上线 | 2026-10 | ai-service + 知识库 + 云端模型 |
| M4 - 自部署就绪 | 2026-11 | Docker Compose + NAS 部署 |

### 7.3 风险与应对

| 风险 | 影响 | 应对策略 |
|------|------|---------|
| NAS 环境多样性 | 兼容性问题 | 提供 ARM/AMD64 多架构镜像 |
| 存储空间不足 | 上传失败 | 配额管理 + 清理策略 |
| 网络不稳定 | 远程访问中断 | 离线模式 + 断点续传 |
| 数据迁移困难 | 升级丢数据 | 迁移脚本 + 自动备份 |
| 安全漏洞 | 数据泄露 | 安全审计 + 依赖更新 |

## 附录：环境变量与端口

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| `JWT_SECRET` | 是 | - | JWT 签名密钥 |
| `POSTGRES_PASSWORD` | 否 | `mvault` | PostgreSQL 密码 |
| `MINIO_ACCESS_KEY` | 否 | `minioadmin` | MinIO 访问密钥 |
| `MINIO_SECRET_KEY` | 否 | `minioadmin` | MinIO 密钥 |
| `OPENAI_API_KEY` | 否 | - | OpenAI API 密钥 |

| 服务 | 内部端口 | 外部端口 |
|------|---------|---------|
| Nginx | 80/443 | 8080/8443 |
| auth-service | 8081 | - |
| backup-service | 8082 | - |
| ai-service | 8083 | - |
| push-service | 8084 | - |
| PostgreSQL | 5432 | - |
| Redis | 6379 | - |
| MinIO | 9000/9001 | 9001 |
| ChromaDB | 8000 | - |