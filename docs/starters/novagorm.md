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

## 禁用 ORM Magic

`novagorm` 把 GORM 作为显式的数据映射与查询工具。`AutoMigrate` 是 Model First 的执行方式，不属于 ORM Magic；Starter 在原生迁移前安装强制门禁，阻止 Model 使用隐式行为。

### 门禁安装范围

以下方式获得的 `*gorm.DB` 都会安装门禁，调用方式仍然是原生 `db.AutoMigrate(...)`：

- `novagorm.DB()`
- `novagorm.Named(name).DB()`
- `novagorm.Register(name, builder)` 返回的自定义连接
- `novagorm.OpenMySQLFromSQLDB(sqlDB)`

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
- 门禁不拦截运行期 CRUD。`Save`、`FullSaveAssociations` 等行为仍按工程规范禁止，更新应使用带明确 `Where` 和更新列的 `Update`/`Updates`。
- Starter 不会自行调用 `AutoMigrate`；应用负责在明确的启动位置执行。

完整工程约束见 [Nova 工程最佳实践：禁用 ORM Magic](../nova_engineering_best_practices.md#禁用-orm-magic)。
