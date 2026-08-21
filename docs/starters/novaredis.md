# novaredis

`starter/cache/novaredis` 负责 Redis 客户端初始化。

## 单实例配置

```yaml
redis:
  addr: localhost:6379
  username: ""
  password: ""
  db: 0
  pool_size: 20
  min_idle_conns: 5
  max_retries: 3
  dial_timeout: 5
  read_timeout: 3
  write_timeout: 3
```

## 多实例配置

```yaml
redis:
  default: main
  instances:
    main:
      addr: localhost:6379
      db: 0
    cache:
      addr: localhost:6380
      db: 1
```

## 最小用法

```go
package main

import "github.com/luaxlou/nova/starter/cache/novaredis"

func main() {
	client, err := novaredis.Client()
	if err != nil {
		panic(err)
	}
	defer client.Close()
}
```

## 约定

- 配置顶层 key 为 `redis`
- 默认客户端使用 `novaredis.Client()` 或 `novaredis.Get().Client()`
- 命名客户端使用 `novaredis.Named("cache").Client()`
