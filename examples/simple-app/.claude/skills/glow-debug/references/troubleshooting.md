# Glow 故障排查指南

## 常见错误

### 连接错误

#### Failed to connect to Glow Server

**错误信息：**
```
Failed to connect to Glow Server: dial tcp 127.0.0.1:32101: connect: connection refused
```

**原因：**
- glow-server 未启动
- 端口被占用
- 防火墙阻止

**解决方案：**

1. 启动 glow-server
```bash
glow-server serve
```

2. 检查端口占用
```bash
lsof -i :32101
lsof -i :32102
```

3. 检查防火墙
```bash
# macOS
sudo pfctl -s rules | grep 32102

# Linux
sudo iptables -L -n | grep 32102
```

4. 使用本地配置降级
```bash
# 创建 local_config.json
cat > local_config.json << 'EOF'
{
  "debug": true,
  "database": {
    "host": "localhost",
    "port": 3306
  }
}
EOF
```

### 配置错误

#### Config not found

**错误信息：**
```
Failed to load config: config not found for app: my-app
```

**原因：**
- 配置文件未生成
- 应用名称不匹配
- 配置文件路径不正确

**解决方案：**

1. 检查应用名称
```bash
# 确保 glowapp.Init() 中的名称一致
glowapp.Init("my-app")  # 必须与 app.yaml 中的名称一致
```

2. 生成配置文件
```bash
# 创建 app.yaml
cat > app.yaml <<EOF
apiVersion: v1
kind: App
metadata:
  name: my-app
spec:
  config:
    debug: true
EOF

# 应用配置生成文件
glow apply -f app.yaml
```

3. 检查配置文件位置
```bash
# 托管运行
ls -la /var/lib/glow-server/apps/my-app/my-app_local_config.json

# 本地调试
ls -la my-app_local_config.json
```

### 数据库错误

#### MySQL connection failed

**错误信息：**
```
MySQL not available: dial tcp 127.0.0.1:3306: connect: connection refused
```

**原因：**
- MySQL 未启动
- 数据库未创建
- 连接信息错误

**解决方案：**

1. 检查 app.yaml 中的配置
```bash
cat app.yaml | grep mysql_dsn
```

2. 检查生成的配置文件
```bash
cat <appName>_local_config.json | grep mysql_dsn
```

3. 测试 MySQL 连接
```bash
# 使用配置文件中的连接信息
mysql -u user -p -h localhost -e "SHOW DATABASES;"
```

4. 确认 MySQL 服务运行正常
```bash
sudo systemctl status mysql
```

5. 检查数据库是否存在
```bash
mysql -u user -p -h localhost -e "USE my_app_db; SHOW TABLES;"
```

#### Redis connection failed

**错误信息：**
```
Redis not available: dial tcp 127.0.0.1:6379: connect: connection refused
```

**解决方案：**

1. 检查 app.yaml 中的配置
```bash
cat app.yaml | grep redis
```

2. 检查生成的配置文件
```bash
cat <appName>_local_config.json | grep redis
```

3. 检查 Redis 状态
```bash
redis-cli ping
```

4. 测试连接
```bash
redis-cli
> PING
PONG
```

### 端口错误

#### Port already in use

**错误信息：**
```
bind: address already in use
```

**原因：**
- 端口被其他进程占用
- 上次应用未正确停止

**解决方案：**

1. 查找占用端口的进程
```bash
lsof -i :8080
```

2. 停止占用端口的进程
```bash
# 如果是 glow 应用
glow stop app my-app

# 如果是其他进程
kill -9 <PID>
```

3. 让 glow-server 自动分配端口
```bash
# 不设置 OP_APP_PORT
unset OP_APP_PORT
```

### 部署错误

#### Upload failed

**错误信息：**
```
Error: Failed to upload application: HTTP 403
```

**原因：**
- API Key 错误
- 认证失败

**解决方案：**

1. 检查 API Key
```bash
glow-server keygen
```

2. 检查 context 配置
```bash
glow context list
glow auth view
```

3. 重置认证
```bash
glow auth reset
# 重新配置 context
```

#### App failed to start

**错误信息：**
```
Error: Application failed to start
```

**解决方案：**

1. 查看应用日志
```bash
glow logs my-app
```

2. 检查应用详情
```bash
glow describe app my-app
```

3. 手动运行应用（调试）
```bash
# 在应用目录
./my-app
```

## 调试技巧

### 启用详细日志

```go
import (
    "log"
    "os"
)

func main() {
    // 设置日志级别
    if os.Getenv("DEBUG") == "true" {
        log.SetFlags(log.LstdFlags | log.Lshortfile)
    }

    glowapp.Init("my-app")
    // ...
}
```

### 调试端点

```go
r := glowhttp.Router()

// 调试信息端点
r.GET("/debug", func(c *gin.Context) {
    c.JSON(200, gin.H{
        "app": glowapp.Name(),
        "env": map[string]string{
            "OP_APP_NAME": os.Getenv("OP_APP_NAME"),
            "OP_APP_PORT": os.Getenv("OP_APP_PORT"),
            "OP_SERVER_URL": os.Getenv("OP_SERVER_URL"),
        },
    })
})

// 健康检查端点
r.GET("/health", func(c *gin.Context) {
    status := map[string]string{
        "app": "ok",
    }

    // 检查数据库
    if db != nil {
        sqlDB, _ := db.DB()
        if err := sqlDB.Ping(); err != nil {
            status["database"] = "error: " + err.Error()
        } else {
            status["database"] = "ok"
        }
    } else {
        status["database"] = "not configured"
    }

    // 检查 Redis
    if client != nil {
        if err := client.Ping(ctx).Err(); err != nil {
            status["redis"] = "error: " + err.Error()
        } else {
            status["redis"] = "ok"
        }
    } else {
        status["redis"] = "not configured"
    }

    c.JSON(200, status)
})
```

### 性能分析

```go
import (
    "net/http"
    "net/http/pprof"
)

r := glowhttp.Router()

// 添加 pprof 端点（仅开发环境）
if os.Getenv("DEBUG") == "true" {
    r.GET("/debug/pprof/", gin.WrapF(pprof.Index))
    r.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
    r.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
    r.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
    r.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))
}
```

使用：
```bash
# CPU profile
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# Memory profile
go tool pprof http://localhost:8080/debug/pprof/heap
```

## 日志分析

### 查看应用日志

```bash
# 实时日志
glow logs my-app

# 最近的日志
glow logs my-app | tail -n 50

# 搜索错误
glow logs my-app | grep -i error
```

### 查看服务器日志

```bash
# 查找 server 进程
ps aux | grep glow-server

# 查看日志文件
tail -f /var/lib/glow-server/logs/glow-server.log  # Linux
tail -f ~/Library/Logs/glow-server.log              # macOS
```

## 性能问题

### 应用响应慢

**排查步骤：**

1. 检查应用状态
```bash
glow describe app my-app
# 查看 CPU 和内存使用
```

2. 查看日志
```bash
glow logs my-app | grep -i "slow\|timeout"
```

3. 性能分析
```bash
# 使用 pprof
go tool pprof http://localhost:8080/debug/pprof/profile
```

4. 数据库查询优化
```bash
# 查看 MySQL 慢查询
mysql -u root -p -e "SHOW VARIABLES LIKE 'slow_query_log';"

# 启用慢查询日志
mysql -u root -p -e "SET GLOBAL slow_query_log = 'ON';"
mysql -u root -p -e "SET GLOBAL long_query_time = 1;"
```

### 内存泄漏

**排查步骤：**

1. 查看内存使用
```bash
glow describe app my-app
# 观察 Memory 使用是否持续增长
```

2. 内存分析
```bash
# 获取 heap profile
curl http://localhost:8080/debug/pprof/heap > heap.prof

# 分析
go tool pprof heap.prof
```

3. goroutine 泄漏检查
```bash
curl http://localhost:8080/debug/pprof/goroutine?debug=2
```

## 网络问题

### DNS 解析问题

```bash
# 检查域名配置
glow get ingress

# 测试 DNS
nslookup myapp.example.com

# 检查 Nginx 配置
cat /var/lib/glow-server/nginx/my-app.conf
```

### 反向代理问题

```bash
# 检查 Nginx 状态
sudo nginx -t
sudo systemctl status nginx

# 查看 Nginx 日志
sudo tail -f /var/log/nginx/error.log

# 重载 Nginx
sudo nginx -s reload
```

## 重置与恢复

### 重置 glow-server

```bash
# 1. 停止 glow-server
sudo systemctl stop glow-server  # Linux
# 或 Ctrl+C (前台运行)

# 2. 备份数据（可选）
cp -r /var/lib/glow-server /backup/glow-server-backup

# 3. 删除数据
rm -rf /var/lib/glow-server/db/
rm -rf /var/lib/glow-server/config/

# 4. 重新启动
glow-server serve
# 会自动重新初始化
```

### 恢复应用

```bash
# 1. 重新编译
go build -o my-app

# 2. 重新部署
glow deploy ./my-app

# 3. 恢复配置
glow config edit my-app
# 粘贴之前的配置
```

## 获取帮助

### 收集诊断信息

```bash
# 创建诊断脚本
cat > diagnose.sh << 'EOF'
#!/bin/bash
echo "=== Glow Diagnostic Info ==="
echo ""
echo "Date: $(date)"
echo ""

echo "=== Glow Server Status ==="
glow-server info
echo ""

echo "=== Application List ==="
glow get apps
echo ""

echo "=== Resource List ==="
glow get resources
echo ""

echo "=== Ingress Rules ==="
glow get ingress
echo ""

echo "=== Context Info ==="
glow context list
glow auth view
echo ""

echo "=== System Info ==="
uname -a
echo ""

echo "=== Disk Usage ==="
df -h
echo ""

echo "=== Memory Usage ==="
free -h  # Linux
vm_stat  # macOS
echo ""
EOF

chmod +x diagnose.sh
./diagnose.sh > diagnostic-output.txt
```

### 提交 Issue

提交 issue 时包含：
1. 诊断信息 (`diagnose.sh` 输出)
2. 相关日志 (`glow logs my-app`)
3. 错误信息完整截图
4. 复现步骤
5. Glow 版本 (`glow --version`)

## 常见问题 FAQ

**Q: glow-server 启动失败？**
A: 检查端口是否被占用，查看错误日志

**Q: 应用无法连接到 glow-server？**
A: 确保 glow-server 运行中，检查防火墙设置

**Q: 配置修改后不生效？**
A: 修改 `app.yaml` 后执行 `glow apply -f app.yaml` 重新生成配置文件，然后使用 `glow restart app my-app` 重启应用

**Q: 数据库连接失败？**
A: 检查 `app.yaml` 中 `spec.config.mysql_dsn` 配置是否正确，确认 MySQL 服务运行正常，数据库已创建

**Q: 端口冲突？**
A: 让 glow-server 自动分配端口，或不设置 OP_APP_PORT

**Q: 如何完全重置 glow？**
A: 删除 glow-server 数据目录后重新启动
