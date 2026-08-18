---
name: glow-deploy
description: Glow 应用部署与运维指南。包含 glow CLI 命令、部署工作流、配置管理、多环境部署、CI/CD 集成、日志查看和故障排查。当用户需要：部署 glow 应用、管理应用生命周期、查看日志、配置多环境、集成 CI/CD 时使用。
---

# Glow Deploy

Glow 应用的部署、运维和管理指南。

## 快速开始

### 配置与部署

```bash
# 1. 编写应用配置文件
cat > app.yaml <<EOF
apiVersion: v1
kind: App
metadata:
  name: my-app
spec:
  binary: ./my-app
  port: 8080
  config:
    mysql_dsn: "user:pass@tcp(localhost:3306)/myapp_db"
    redis_addr: "localhost:6379"
EOF

# 2. 应用配置
glow apply -f app.yaml

# 3. 启动应用
glow start app my-app

# 4. 查看应用状态
glow get apps

# 5. 查看日志
glow logs my-app
```

## Glow CLI 命令

### 应用配置（核心命令）

```bash
# 应用配置文件（唯一资源配置方式）
glow apply -f app.yaml

# 功能：
# - 注册/更新应用元数据
# - 配置 Ingress（域名绑定）
# - 生成应用配置文件
```

**详细文档**: 参见 [Glow Apply 手册](../../docs/glow_apply_manual.md)

### 资源查看

```bash
# 列出所有应用
glow get apps

# 列出所有网关路由
glow get ingress

# 查看节点信息
glow get nodes

# 查看基础设施资源
glow get resources

# 查看应用详情
glow describe app my-app

# 查看节点详情
glow describe node localhost
```

### 日志查看

```bash
# 查看应用日志
glow logs my-app

# 实时跟踪日志
glow logs my-app -f
```

### 生命周期管理

```bash
# 启动应用
glow start app my-app

# 停止应用
glow stop app my-app

# 重启应用
glow restart app my-app

# 删除应用
glow delete app my-app

# 删除路由
glow delete ingress my-app
```

### 配置管理

Glow 采用**完全声明式的配置管理**。所有应用配置通过 `app.yaml` 声明，使用 `glow apply` 应用。

```bash
# 配置声明：在 app.yaml 的 spec.config 字段中声明配置
# 应用配置：执行 glow apply -f app.yaml
# 查看配置：读取生成的配置文件
cat /var/lib/glow-server/apps/my-app/my-app_local_config.json
```

**配置管理原则**:
- ✅ 配置即代码：所有配置在 YAML 文件中管理
- ✅ 版本控制：配置变更可通过 Git 追踪
- ✅ 不可变性：不支持运行时通过命令行修改配置
- ❌ 已移除：`glow config set/get/list/export/edit` 等命令

### 环境管理

```bash
# 列出所有环境
glow context list

# 切换环境
glow context use prod

# 添加环境
glow context add prod --url http://prod-server:32102 --key <api-key>

# 查看认证信息
glow auth view

# 重置认证
glow auth reset
```

## 部署工作流

### 标准流程

```bash
# 1. 编写应用配置文件
vim app.yaml

# 2. 应用配置
glow apply -f app.yaml

# 3. 启动应用
glow start app my-app

# 4. 验证
glow get apps
glow logs my-app
```

### 更新应用配置

```bash
# 1. 修改配置文件
vim app.yaml

# 2. 重新应用配置
glow apply -f app.yaml

# 3. 重启应用使新配置生效
glow restart app my-app
```

## 配置管理

### 配置即代码

所有配置在 `app.yaml` 的 `spec.config` 字段中声明：

```yaml
apiVersion: v1
kind: App
metadata:
  name: my-app
spec:
  config:
    debug: false
    mysql_dsn: "user:pass@tcp(prod-db:3306)/prod_db"
    redis_addr: "prod-redis.example.com:6379"
    log_level: "info"
```

### 配置文件生成

执行 `glow apply -f app.yaml` 后，配置会写入：
```
<data-dir>/apps/<app-name>/<app-name>_local_config.json
```

### 环境特定配置

为不同环境创建不同的 YAML 文件：

```bash
app.yaml              # 开发环境
app-production.yaml   # 生产环境
app-staging.yaml      # 测试环境
```

## 多环境管理

### 环境配置

```bash
# 开发环境
glow context add dev --url http://dev-server:32102 --key <dev-key>

# 生产环境
glow context add prod --url http://prod-server:32102 --key <prod-key>

# 测试环境
glow context add test --url http://test-server:32102 --key <test-key>
```

### 环境切换

```bash
# 查看当前环境
glow context list

# 切换到生产环境
glow context use prod

# 部署到生产环境
glow deploy ./my-app
```

## CI/CD 集成

### GitHub Actions

```yaml
name: Deploy to Glow

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build
        run: go build -o app

      - name: Deploy
        env:
          GLOW_SERVER_URL: ${{ secrets.GLOW_SERVER_URL }}
          GLOW_API_KEY: ${{ secrets.GLOW_API_KEY }}
        run: |
          curl -fsSL "https://raw.githubusercontent.com/luaxlou/glow/main/scripts/install-local.sh" | bash
          glow context add production --server-url $GLOW_SERVER_URL --api-key $GLOW_API_KEY
          glow context use production
          glow apply -f app.yaml
          glow start app my-app
```

### GitLab CI

```yaml
deploy:
  stage: deploy
  image: golang:1.21
  script:
    - go build -o app
    - curl -fsSL "https://raw.githubusercontent.com/luaxlou/glow/main/scripts/install-local.sh" | bash
    - glow context add production --server-url $GLOW_SERVER_URL --api-key $GLOW_API_KEY
    - glow context use production
    - glow apply -f app.yaml
    - glow start app my-app
  only:
    - main
```

## 域名配置

### 通过 API 配置

```bash
curl -H "Authorization: Bearer <api-key>" \
  -H "Content-Type: application/json" \
  -d '{"app_name":"my-app","domain":"myapp.example.com","port":8080}' \
  http://localhost:32102/ingress/update
```

### Nginx 配置

确保主配置包含：

```nginx
include /var/lib/glow-server/nginx/*.conf;
```

## 日志管理

### 查看日志

```bash
# 实时日志
glow logs my-app

# 日志位置
# Linux: /var/lib/glow-server/apps/<app-name>/logs/<app-name>.log
# macOS: ~/Library/Application Support/glow-server/apps/<app-name>/logs/
```

### 日志轮转

- 单文件最大 10MB
- 保留最近 5 个文件

## 故障排查

### 应用无法启动

```bash
# 1. 查看 glow-server 日志
glow-server info

# 2. 查看应用日志
glow logs my-app

# 3. 检查应用状态
glow describe app my-app

# 4. 检查端口占用
lsof -i :<port>
```

### 配置未生效

```bash
# 1. 确认配置文件已生成
cat /var/lib/glow-server/apps/my-app/my-app_local_config.json

# 2. 检查 YAML 配置
cat app.yaml

# 3. 重新应用配置
glow apply -f app.yaml

# 4. 重启应用使新配置生效
glow restart app my-app

# 5. 检查应用日志
glow logs my-app
```

### 数据库连接失败

```bash
# 1. 检查 app.yaml 中的配置
cat app.yaml | grep mysql_dsn

# 2. 检查生成的配置文件
cat /var/lib/glow-server/apps/my-app/my-app_local_config.json | grep mysql_dsn

# 3. 测试数据库连接
mysql -u user -p -h localhost -e "SHOW DATABASES;"

# 4. 确认 MySQL 服务运行正常
sudo systemctl status mysql
```

## 性能优化

### 减小二进制大小

```bash
# 编译优化
go build -ldflags="-s -w" -o my-app

# 使用 upx 压缩（可选）
upx --best --lzma my-app
```

### 回滚策略

```bash
# 部署前备份配置
cp app.yaml app.yaml.backup

# 部署新版本
glow apply -f app.yaml
glow start app my-app

# 如有问题，回滚配置
glow apply -f app.yaml.backup
glow restart app my-app
```

## 最佳实践

1. **配置即代码**: 所有配置在 `app.yaml` 中管理，纳入版本控制
2. **声明式配置**: 使用 `glow apply -f app.yaml` 统一管理所有配置
3. **多环境隔离**: 为不同环境创建不同的 YAML 文件，使用 context 切换
4. **日志监控**: 定期查看 `glow logs`
5. **优雅重启**: 使用 `glow restart` 而非 `stop` + `start`
6. **版本管理**: 配置变更通过 Git 追踪，部署前保留备份

## HTTP API 参考

详见 [API Reference](references/http-api.md)

## 常见问题

### Q: 如何查看应用的完整信息？
A: 使用 `glow describe app <name>`

### Q: 配置修改后需要重启吗？
A: 修改 `app.yaml` 后执行 `glow apply -f app.yaml` 重新生成配置文件，然后使用 `glow restart app <name>` 重启应用使新配置生效。

### Q: 如何同时部署多个应用？
A: 为每个应用创建独立的 `app.yaml` 文件，分别执行 `glow apply -f <app>.yaml` 和 `glow start app <name>`
