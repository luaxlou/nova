# novagorm

`starter/gorm/novagorm` 是 GORM Starter。它负责按配置创建一个或多个 `*gorm.DB`，并把具体数据库配置挂在所选 driver 下。

## 多实例配置

```yaml
gorm:
  main:
    driver: mysql
    mysql:
      dsn: root:password@tcp(localhost:3306)/app?parseTime=true
      max_open: 20
      max_idle: 10
      conn_max_lifetime: 300
      conn_max_idle_time: 60
      skip_initialize_with_version: true
  analytics:
    driver: mysql
    mysql:
      dsn: analytics:password@tcp(localhost:3306)/analytics?parseTime=true
```

## 单实例简写

```yaml
gorm:
  driver: mysql
  mysql:
    dsn: root:password@tcp(localhost:3306)/app?parseTime=true
    max_open: 20
    max_idle: 10
```

## 最小用法

```go
package main

import "github.com/luaxlou/nova/starter/gorm/novagorm"

func main() {
	db, err := novagorm.DB()
	if err != nil {
		panic(err)
	}

	_ = db
}
```

## 多实例用法

```go
mainDB, err := novagorm.Named("main").DB()
if err != nil {
	panic(err)
}

analyticsDB, err := novagorm.Named("analytics").DB()
if err != nil {
	panic(err)
}

_, _ = mainDB, analyticsDB
```

## 约定

- 包路径为 `github.com/luaxlou/nova/starter/gorm/novagorm`
- 配置顶层 key 为 `gorm`
- 数据库类型由 `gorm.driver` 或 `gorm.<name>.driver` 选择
- MySQL 配置挂在 `gorm.mysql` 或 `gorm.<name>.mysql` 下
- 多实例通过 `gorm.<name>` 配置
- 指定实例使用 `novagorm.Named("analytics").DB()`
- 只有一个实例时可以使用 `novagorm.DB()`
- 有多个实例时必须使用 `novagorm.Named("<name>").DB()`
- MySQL 不作为独立 Starter 对外提供；使用 MySQL 时通过 `gorm.driver: mysql` 与 `gorm.mysql` 配置
- 当前内置配置支持 MySQL driver；其他 driver 可通过 `Register` 扩展
