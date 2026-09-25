# novagorm

`starter/gorm/novagorm` 是 GORM Starter。它负责按配置创建一个或多个 `*gorm.DB`，内置 `mysql` 与 `postgres` driver，并把具体数据库配置挂在所选 driver 下。PostgreSQL 的真实验证基线为 PostgreSQL 18。

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

PostgreSQL 单实例使用固定 driver ID `postgres`：

```yaml
gorm:
  driver: postgres
  postgres:
    dsn: host=127.0.0.1 user=app password=<runtime-secret> dbname=app port=5432 sslmode=disable TimeZone=UTC
    max_open: 20
    max_idle: 10
    conn_max_lifetime: 1800
    conn_max_idle_time: 300
    prefer_simple_protocol: false
```

MySQL 与 PostgreSQL 可以作为命名实例并存；`default` 用于选择 `novagorm.DB()` 返回的实例：

```yaml
gorm:
  default: main
  main:
    driver: postgres
    postgres:
      dsn: host=127.0.0.1 user=app password=<runtime-secret> dbname=app port=5432 sslmode=disable
  legacy:
    driver: mysql
    mysql:
      dsn: root:<runtime-secret>@tcp(localhost:3306)/legacy?parseTime=true
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
- PostgreSQL 配置挂在 `gorm.postgres` 或 `gorm.<name>.postgres` 下；不支持 `postgresql` 别名
- PostgreSQL 的 `prefer_simple_protocol` 用于关闭 pgx 隐式 prepared statement；默认值为 `false`
- 两种 driver 都支持 `max_open`、`max_idle`、`conn_max_lifetime` 和 `conn_max_idle_time`；小于等于零时保留 Go SQL 默认值
- 多实例通过 `gorm.<name>` 配置
- 指定实例使用 `novagorm.Named("analytics").DB()`
- 只有一个实例时可以使用 `novagorm.DB()`
- 有多个实例时必须使用 `novagorm.Named("<name>").DB()`
- MySQL 与 PostgreSQL 都是 `novagorm` 的内置 driver，不作为独立 Starter 对外提供；其他 driver 可通过 `Register` 扩展
- DSN 与密码必须由运行时配置或密钥管理系统提供，不得写入日志、错误信息或提交到仓库

## 从现有 SQL 连接打开 GORM

已有 `*sql.DB` 时，可以通过显式 bridge 接入同一 Model First 门禁：

```go
mysqlDB, err := novagorm.OpenMySQLFromSQLDB(mysqlSQLDB)
postgresDB, err := novagorm.OpenPostgresFromSQLDB(postgresSQLDB)
```

bridge 不接管 DSN 解析，但 `Close` / `CloseAll` 仍会关闭底层连接池。

## 禁用 ORM Magic

`novagorm` 把 GORM 作为显式的数据映射与查询工具。`AutoMigrate` 是 Model First 的执行方式，不属于 ORM Magic；Starter 在原生迁移前安装强制门禁，阻止 Model 使用隐式行为。

### 门禁安装范围

以下方式获得的 `*gorm.DB` 都会安装门禁，调用方式仍然是原生 `db.AutoMigrate(...)`：

- `novagorm.DB()`
- `novagorm.Named(name).DB()`
- `novagorm.Register(name, builder)` 返回的自定义连接
- `novagorm.OpenMySQLFromSQLDB(sqlDB)`
- `novagorm.OpenPostgresFromSQLDB(sqlDB)`

直接调用 `gorm.Open` 创建的连接不由 Nova 管理，不会安装门禁。

### 拦截规则

调用 `AutoMigrate` 时会先解析并检查本次传入的全部 Model。以下任一情况都会返回 `novagorm.ErrORMMagic`：

- 缺少 `TableName() string`。
- 持久化字段缺少 `gorm:"column:..."`；`gorm:"-"` 字段除外。
- 字段启用了 `autoCreateTime` 或 `autoUpdateTime`。常规 `CreatedAt`、`UpdatedAt` 必须分别声明 `autoCreateTime:false`、`autoUpdateTime:false`。
- 嵌入 `gorm.Model`。
- 使用 `gorm.DeletedAt`。
- 定义 GORM lifecycle hook：`BeforeCreate`、`AfterCreate`、`BeforeUpdate`、`AfterUpdate`、`BeforeSave`、`AfterSave`、`BeforeDelete`、`AfterDelete` 或 `AfterFind`。
- 声明 GORM association 字段；跨表关系应保留显式外键 ID，并由 data capability 查询或写入。

门禁先校验全部 Model，再调用原生 Migrator；只要一个 Model 不合规，本次 `AutoMigrate` 就不会执行任何 DDL：

```go
err := db.AutoMigrate(
	&userdata.UserModel{},
	&orderdata.OrderModel{},
)
if errors.Is(err, novagorm.ErrORMMagic) {
	return fmt.Errorf("GORM model rejected: %w", err)
}
```

典型错误包含 Model、字段或规则：

```text
novagorm: ORM magic blocked for model UserModel: field UpdatedAt enables autoUpdateTime
```

### 边界

- 门禁只检查本次传入 `AutoMigrate` 的 Model。所有持久化 Model 必须明确列入迁移清单。
- 门禁不维护 SQL migration；Schema 的唯一来源仍然是 Model。
- Model 只保留显式外键 ID，不声明 GORM association。因为 GORM 依赖 association 元数据生成数据库外键，Nova 不自动生成外键约束；消费项目必须单独设计和验证需要的数据库外键。
- Nova 不提供 MySQL 到 PostgreSQL 的数据迁移工具。
- 门禁不拦截运行期 CRUD。`Save`、`FullSaveAssociations` 等行为仍按工程规范禁止，更新应使用带明确 `Where` 和更新列的 `Update`/`Updates`。
- Starter 不会自行调用 `AutoMigrate`；应用负责在明确的启动位置执行。

完整工程约束见 [Nova 工程最佳实践：禁用 ORM Magic](../nova_engineering_best_practices.md#禁用-orm-magic)。
