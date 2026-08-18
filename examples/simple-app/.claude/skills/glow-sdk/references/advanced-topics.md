# Glow SDK 高级主题

## 自定义 Starter

创建自己的 starter 组件：

```go
package mystarter

import (
    "github.com/luaxlou/glow/starter/glowapp"
    "github.com/luaxlou/glow/starter/glowapp/config"
)

type MyComponent struct {
    config Config
}

type Config struct {
    Enabled bool   `json:"enabled"`
    Host    string `json:"host"`
    Port    int    `json:"port"`
}

func Init() (*MyComponent, error) {
    // 加载配置
    var cfg Config
    if err := config.Get("mycomponent", &cfg); err != nil {
        return nil, err
    }

    component := &MyComponent{config: cfg}

    // 注册清理函数
    glowapp.RegisterCleanup("mycomponent", func() {
        component.Close()
    })

    return component, nil
}

func (c *MyComponent) Close() {
    // 清理逻辑
}
```

## 中间件集成

### Gin 中间件

```go
r := glowhttp.Router()

// 添加中间件
r.Use(gin.Logger())
r.Use(gin.Recovery())

// 自定义中间件
r.Use(func(c *gin.Context) {
    c.Set("app", glowapp.Name())
    c.Next()
})
```

## 错误处理

### 全局错误处理

```go
r := glowhttp.Router()

// 自定义错误处理
r.Use(func(c *gin.Context) {
    c.Next()

    if len(c.Errors) > 0 {
        err := c.Errors.Last()
        c.JSON(500, gin.H{
            "error": err.Error(),
        })
    }
})
```

## 健康检查

```go
r := glowhttp.Router()

// 健康检查端点
r.GET("/health", func(c *gin.Context) {
    // 检查数据库
    if db != nil {
        sqlDB, _ := db.DB()
        if err := sqlDB.Ping(); err != nil {
            c.JSON(503, gin.H{
                "status": "unhealthy",
                "database": "down",
            })
            return
        }
    }

    // 检查 Redis
    if client != nil {
        if err := client.Ping(ctx).Err(); err != nil {
            c.JSON(503, gin.H{
                "status": "unhealthy",
                "redis": "down",
            })
            return
        }
    }

    c.JSON(200, gin.H{
        "status": "healthy",
        "app": glowapp.Name(),
    })
})
```

## 优雅停机

### 数据库连接池关闭

```go
glowapp.RegisterCleanup("database", func() {
    if db != nil {
        sqlDB, _ := db.DB()
        sqlDB.Close()
        log.Println("Database connection closed")
    }
})
```

### HTTP 连接清理

```go
glowapp.RegisterCleanup("http-client", func() {
    if httpClient != nil {
        httpClient.CloseIdleConnections()
        log.Println("HTTP client connections closed")
    }
})
```

## 性能优化

### 数据库连接池配置

```go
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(100)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(time.Hour)
```

### Redis 连接池配置

```go
client := glowredis.Client()
client.Options().PoolSize = 10
client.Options().MinIdleConns = 5
```

## 监控指标

### 自定义指标

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path"},
    )
)

func init() {
    prometheus.MustRegister(requestsTotal)
}

r := glowhttp.Router()
r.Use(func(c *gin.Context) {
    requestsTotal.WithLabelValues(c.Request.Method, c.FullPath()).Inc()
    c.Next()
})

// 指标端点
r.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

## 测试技巧

### Mock Glow Server

```go
func TestWithMockServer(t *testing.T) {
    // 启动测试 glow-server
    cmd := exec.Command("glow-server", "serve")
    if err := cmd.Start(); err != nil {
        t.Fatalf("Failed to start glow-server: %v", err)
    }
    defer cmd.Process.Kill()

    // 等待启动
    time.Sleep(2 * time.Second)

    // 运行测试
    glowapp.Init("test-app")
    // ...
}
```

### 表格驱动测试

```go
func TestConfigLoading(t *testing.T) {
    tests := []struct {
        name    string
        config  string
        want    Config
        wantErr bool
    }{
        {
            name: "valid config",
            config: `{"debug": true}`,
            want: Config{Debug: true},
            wantErr: false,
        },
        // 更多测试用例...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试逻辑
        })
    }
}
```

## 常见模式

### Singleton 模式

```go
var (
    instance *MyService
    once     sync.Once
)

func MyService() (*MyService, error) {
    once.Do(func() {
        instance = &MyService{}
        // 初始化...
    })
    return instance, nil
}
```

### Factory 模式

```go
func NewService(name string) (*Service, error) {
    // 根据名称创建不同的服务
    switch name {
    case "mysql":
        return newMySQLService()
    case "redis":
        return newRedisService()
    default:
        return nil, fmt.Errorf("unknown service: %s", name)
    }
}
```

### Observer 模式

```go
type EventListener interface {
    OnEvent(event Event)
}

type EventEmitter struct {
    listeners []EventListener
}

func (e *EventEmitter) Subscribe(listener EventListener) {
    e.listeners = append(e.listeners, listener)
}

func (e *EventEmitter) Emit(event Event) {
    for _, listener := range e.listeners {
        listener.OnEvent(event)
    }
}
```
