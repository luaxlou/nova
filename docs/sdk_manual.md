## 仓库职责说明

本手册仅覆盖 **nova 框架仓**（starter/SDK）。

# Nova SDK 用户手册

`nova/starter` 是 Nova 框架的 Go SDK，提供应用内能力接入组件。

## 1. 安装

```bash
go get github.com/luaxlou/nova/starter
```

## 2. 快速开始

```go
package main

import (
    "fmt"
    "github.com/luaxlou/nova/starter/novaconfig"
)

func main() {
    logLevel := novaconfig.GetString("log_level")
    maxConnections := novaconfig.GetInt("max_connections")

    fmt.Printf("Log Level: %s\n", logLevel)
    fmt.Printf("Max Connections: %d\n", maxConnections)
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
log_level: debug
max_connections: 100
mysql:
  dsn: root:password@tcp(localhost:3306)/app
redis:
  addr: localhost:6379
```

### 3.2 novagin

提供 HTTP 服务器启动适配（Gin）。

- `Init(port int)`
- `Router() *gin.Engine`
- `Run()`

### 3.3 novamysql

提供 MySQL 连接初始化能力。

- `Init(name string)`
- `Gorm() (*gorm.DB, error)`
- `DB() (*sql.DB, error)`
- `Reload()`

### 3.4 novaredis

提供 Redis 客户端初始化能力。

- `Init()`
- `Client() (*redis.Client, error)`
- `Reload()`

## 4. 运行模式

### 4.1 本地开发

直接在应用目录准备 `config.yaml` 并运行应用。
