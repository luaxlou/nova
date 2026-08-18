# Glow 部署架构

## 架构概览

```
+---------------------------------------------------------------------------------------+
|  HOST MACHINE                                                                         |
|                                                                                       |
|   +-------------------------------------------------------------------------------+   |
|   |  GLOW SERVER (Daemon)                                                         |   |
|   |                                                                               |   |
|   |   [ Process Mgr ] --------+               [ HTTP API ] <---> [ Config DB ]    |   |
|   |          |                |                    ^   |                              |
|   |          | (5) Monitor    | (6) Config         |   +---> [ Provisioner ]      |   |
|   |          v                v                    |               |              |   |
|   +----------|----------------|--------------------|---------------|--------------+   |
|              |                |                    |               | (2) Create       |
|              |                |              (1)   |               |     User/DB      |
|              |                |            Register|               |                  |
|   +----------|-------+   +----|----------+ & Req   |      +--------v-----------+      |
|   | USER APPLICATION |   | NGINX GATEWAY | --------+      | LOCAL INFRA        |      |
|   |                  |   |               |                |                    |      |
|   |  [ Biz Logic ]   |   |  [ Vhost ] <------- Traffic    |  [ MySQL / Redis ] |      |
|   |        |         |   +---------------+                |                    |      |
|   |        v         |                                    |                    |      |
|   |   [ Glow SDK ] <-----------------------------------------+                 |      |
|   |                  |        (4) Connect                 |                    |      |
|   +------------------+                                    +--------------------+      |
|                                                                                       |
+---------------------------------------------------------------------------------------+
```

## 组件说明

### Glow Server

核心守护进程，提供：
- **进程托管**: 替代 Systemd，管理应用生命周期
- **配置中心**: 基于 SQLite 的轻量级配置存储
- **网关自动化**: 自动生成 Nginx 配置
- **资源管理**: 自动申请和配置 MySQL、Redis 等

### 应用 SDK

集成在应用中的 SDK，提供：
- **服务注册**: 向 Glow Server 注册应用身份
- **配置获取**: 自动拉取配置，支持热更新
- **资源连接**: 自动连接基础设施资源

## 数据流

### 应用启动流程

1. **配置阶段**: `glow apply -f app.yaml`
   - 注册应用元数据
   - 配置 Ingress（域名绑定）
   - 生成应用配置文件（`<appName>_local_config.json`）
2. **启动阶段**: `glow start app <name>`
   - Server 设置环境变量（`OP_APP_NAME`, `OP_APP_PORT`）
   - 启动应用进程
3. **运行阶段**: 应用启动
   - SDK 初始化: `glowapp.Init("app-name")`
   - 配置读取: SDK 从本地配置文件读取配置
   - 资源连接: 使用配置中的数据库连接等信息
   - 启动 HTTP: SDK 启动 HTTP 服务，监听分配的端口
4. **网关配置**: Server 自动配置 Nginx 反向代理（如果指定了 domain）

### 配置更新流程

1. **修改配置**: 编辑 `app.yaml` 中的 `spec.config` 字段
2. **应用配置**: `glow apply -f app.yaml`
   - 重新生成配置文件（`<appName>_local_config.json`）
   - 如果应用正在运行且配置变化，自动重启应用
3. **配置生效**: 应用重启后从新配置文件读取配置

### 应用监控流程

1. **进程监控**: Server 监控应用进程状态
2. **自动重启**: 应用崩溃时，Server 自动重启
3. **日志轮转**: Server 自动管理应用日志
4. **状态查询**: 通过 `glow get apps` 和 `glow describe app <name>` 查看应用状态

## 部署模式

### 单机部署

```
Server (localhost)
  ├── App 1 (port 54321)
  ├── App 2 (port 54322)
  └── App 3 (port 54323)

Nginx
  ├── app1.example.com → App 1
  ├── app2.example.com → App 2
  └── app3.example.com → App 3
```

### 多机部署

```
Server 1 (prod-server-1)
  ├── App A (instance 1)
  └── App B (instance 1)

Server 2 (prod-server-2)
  ├── App A (instance 2)
  └── App B (instance 2)

Load Balancer
  ├── App A traffic → Server 1, Server 2
  └── App B traffic → Server 1, Server 2
```

## 资源管理

### MySQL 配置

```yaml
# app.yaml
spec:
  config:
    mysql_dsn: "user:pass@tcp(localhost:3306)/my_app_db"
```

```bash
# 应用配置
glow apply -f app.yaml

# 生成的配置文件包含 mysql_dsn
# 应用从配置文件读取连接信息
```

### Redis 配置

```yaml
# app.yaml
spec:
  config:
    redis_addr: "localhost:6379"
    redis_password: ""
    redis_db: 0
```

```bash
# 应用配置
glow apply -f app.yaml

# 生成的配置文件包含 Redis 连接信息
# 应用从配置文件读取连接信息
```

## 安全考虑

### API 认证

所有 HTTP API 需要认证：

```http
Authorization: Bearer <api_key>
```

API Key 生成：
```bash
glow-server keygen
```

### 网络隔离

- **HTTP API** (32102): 仅本地访问，或通过防火墙限制
- **App Center** (32101): 仅本地访问
- **应用端口**: 自动分配，通过 Nginx 暴露

### 配置隔离

- 每个应用有独立的配置文件
- 不同应用使用不同的数据库连接配置
- 应用间通过配置隔离

## 扩展性

### 水平扩展

```bash
# 在多台服务器上部署相同的 glow-server
# 每台服务器运行独立的应用实例

# 使用外部负载均衡器分发流量
```

### 垂直扩展

```bash
# 增加服务器资源
# 在同一台服务器上运行更多应用实例

# 使用 glow deploy --name app-name-1
# 使用 glow deploy --name app-name-2
```

## 高可用

### 进程保活

- 应用崩溃自动重启
- 指数退避策略防止频繁重启
- 最大重启次数限制

### 配置备份

```bash
# 备份配置文件（推荐）
git add app.yaml
git commit -m "Backup app configuration"

# 备份生成的配置文件
cp /var/lib/glow-server/apps/<app-name>/<app-name>_local_config.json /backup/
```

### 日志持久化

```bash
# 日志文件位置
/var/lib/glow-server/apps/<app-name>/logs/<app-name>.log

# 日志轮转
- 单文件最大 10MB
- 保留最近 5 个文件
```

## 监控与告警

### 应用监控

```bash
# 查看应用状态
glow describe app my-app

# 输出：
# - PID
# - 端口
# - CPU 使用率
# - 内存使用
# - 运行时间
# - 重启次数
```

### Server 监控

```bash
# 查看 Server 状态
glow-server info

# 输出：
# - 运行状态
# - 资源配置
# - 应用列表
```

## 故障恢复

### 应用崩溃

1. Server 检测到应用退出
2. 等待退避时间（1s, 2s, 4s, ...）
3. 自动重启应用
4. 记录崩溃日志

### Server 崩溃

1. 应用检测到 Server 连接断开
2. 降级使用本地配置
3. 等待 Server 恢复
4. 自动重连

### 数据库故障

1. 应用检测到数据库连接失败
2. 记录错误日志
3. 可选：降级到缓存模式
4. 等待数据库恢复

## 最佳实践

1. **配置即代码**: 所有配置在 `app.yaml` 中管理，纳入版本控制
2. **资源隔离**: 每个应用使用独立的数据库连接配置
3. **声明式配置**: 使用 `glow apply -f app.yaml` 统一管理配置
4. **优雅停机**: 使用 `glow stop` 而非 `kill -9`
5. **日志管理**: 定期检查和清理日志
6. **监控告警**: 设置应用重启次数告警
7. **备份策略**: 配置文件纳入 Git 版本控制
