## 仓库职责说明

本手册仅覆盖 **nova 框架仓**（starter/SDK）。

# Nova SDK 用户手册

`nova` 是 Nova 框架的 Go SDK，提供应用内能力接入组件、Starter 与 ORM 工具。

## 1. 安装

```bash
go get github.com/luaxlou/nova
```

## 2. 快速开始

```go
package main

import (
    "fmt"

    "github.com/luaxlou/nova/starter/config/novaconfig"
)

func main() {
    fmt.Println(novaconfig.GetString("app.name"))
}
```

## 3. 核心组件

### 3.1 novaconfig

从本地配置文件读取配置。

- `Get(key string) any`
- `GetString(key string) string`
- `GetInt(key string) int`
- `GetBool(key string) bool`
- `Reload() error`
- `SetConfigPath(path string)`

默认读取 `config.yaml`。

示例：

```yaml
app:
  name: demo

# starter/http/novagin
http:
  port: 8080

# starter/gorm/novagorm
gorm:
  main:
    driver: mysql
    mysql:
      dsn: root:password@tcp(localhost:3306)/app
      max_open: 20
      max_idle: 10
      skip_initialize_with_version: true
  analytics:
    driver: mysql
    mysql:
      dsn: analytics:password@tcp(localhost:3306)/analytics

# starter/cache/novaredis
redis:
  main:
    addr: localhost:6379
    db: 0
  cache:
    addr: localhost:6380
    db: 1
```

### 3.2 novagin

提供 HTTP 服务器启动适配（Gin）。

- `Router() *gin.Engine`
- `Run()`

### 3.3 novagorm

提供 GORM 动态装配能力。`novagorm` 位于 `starter/gorm/novagorm`，是 GORM Starter。GORM 支持多实例；数据库类型由 `gorm.<name>.driver` 选择，使用 MySQL 时配置放在 `gorm.<name>.mysql` 下。有多个实例时必须使用 `Named(name)` 获取。

- `Register(name string, builder Builder)`
- `OpenMySQLFromSQLDB(db *sql.DB) (*gorm.DB, error)`
- `Get() *gormInstance`
- `Named(name string) *gormInstance`
- `DB() (*gorm.DB, error)`
- `Reload()`
- `Close() error`
- `CloseAll() error`

### 3.4 novaredis

提供 Redis 客户端初始化能力。

- `Client() (*redis.Client, error)`
- `Get() *redisInstance`
- `Named(name string) *redisInstance`
- `Reload()`
- `Close() error`
- `CloseAll() error`

## 4. 运行模式

### 4.1 本地开发

直接在应用目录准备 `config.yaml` 并运行应用。
