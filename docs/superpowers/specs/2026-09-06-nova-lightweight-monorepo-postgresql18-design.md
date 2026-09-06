# Nova 轻量单仓与 PostgreSQL 18 Driver 设计

> 日期：2026-09-06
> 状态：规划项，暂不进入实施
> 决策范围：Nova 仓库边界、Starter 索引、`novagorm` PostgreSQL 支持
> 首个消费者：软件供应链安全项目

## 0. 当前处置

本文档只记录已经确认的技术方向和未来实施边界，不代表 PostgreSQL driver 已进入开发。PostgreSQL 18 支持只是软件供应链安全整体建设计划中的一项基础设施能力，不能替代存储模型、运行时模型、数据治理、SBOM 建设和本体水化等整体设计。

在以下前置条件全部满足前，不编写 PostgreSQL driver 实施计划，不修改 Nova 依赖或代码，也不启动软件供应链安全项目的数据库迁移：

1. 软件供应链安全平台整体存储架构完成设计复核。
2. SBOM、不可变快照、内容寻址对象、七对象本体、场景对象和 HITL 状态的持久化边界完成冻结。
3. PostgreSQL 18 在整体建设分期中的优先级、输入、输出和验收责任得到确认。
4. Nova 变更与消费项目迁移被拆分为独立、可审阅的实施阶段。

## 1. 决策摘要

Nova 继续采用单仓库、单 Go Module。所有 Starter 实现保留在 Nova 项目内，不拆成独立仓库或细粒度 Module。Nova 的轻量化通过清晰包边界、按需导入、延迟初始化和受控依赖实现，不通过物理拆仓实现。

PostgreSQL 18 作为 `starter/gorm/novagorm` 的第二个内置 driver，与 MySQL 并列。业务项目继续只依赖 `novagorm.DB()`、`novagorm.Named(name).DB()`、`Reload()` 和 `CloseAll()` 等稳定能力，不自行装配数据库连接生命周期。

Nova 增加机器可读的 Starter 索引。该索引用于发现、文档导航、兼容性检查和 AI Coding 上下文，不是运行时插件注册表，不自动引入、下载或初始化任何 Starter。

## 2. 目标与非目标

### 2.1 目标

- 在不扩张业务抽象的前提下，让 `novagorm` 原生支持 PostgreSQL 18。
- 保持 MySQL 现有配置、公共 API、多实例、手工 `Register` 和关闭语义兼容。
- 对 MySQL 与 PostgreSQL 应用同一套 Model First / ORM Magic 门禁。
- 让开发者和 AI 能从统一索引判断 Nova 当前有哪些 Starter、driver、配置入口、成熟度和文档。
- 约束 Starter 之间的依赖方向，防止 Nova 因能力增加而形成默认全量技术栈。

### 2.2 非目标

- 不建立运行时插件市场、动态加载器或依赖注入容器。
- 不把每个 driver、云产品或 SDK 拆成独立仓库。
- 不把 PostgreSQL 18 专属业务表、索引或查询放进 Nova。
- 不让 `novagorm` 自动执行应用 Model 初始化、健康检查或业务迁移。
- 不在本次设计中迁移软件供应链安全项目的数据和业务模型。
- 不移除 MySQL；是否退役由各消费项目自行决定。

## 3. 单仓轻量化边界

### 3.1 仓库结构

Nova 保持现有按能力组织的目录，并对云厂商能力集中归类：

```text
nova/
├── internal/registry/                 # Starter 实例注册基础
├── starter/
│   ├── config/novaconfig/             # 统一配置入口
│   ├── gorm/novagorm/                 # GORM + MySQL/PostgreSQL drivers
│   ├── cache/novaredis/               # Redis
│   ├── http/novagin/                  # Gin
│   ├── realtime/novawebsocket/        # WebSocket
│   └── aliyun/
│       ├── novaoss/                   # OSS
│       └── novaqwen/                  # Qwen；从现有 ai 目录归入阿里云能力族
├── docs/starters/                     # Starter 专用说明
└── starter-index.yaml                 # 机器可读索引
```

`novaqwen` 的物理迁移是独立兼容性任务。本次 PostgreSQL driver 不顺带移动包路径，避免把目录治理和数据库能力耦合为一次破坏性变更。

### 3.2 “轻量”的工程定义

Nova 轻量化由以下可检查规则保证：

1. 不提供聚合所有 Starter 的总入口包。
2. 应用只编译其显式 import 的 Starter；未配置、未调用的 Starter 不初始化资源。
3. Starter 只依赖 `novaconfig`、内部注册基础和自身官方 SDK；不得横向依赖无关 Starter。
4. 新能力必须归入既有能力族；只有形成新的基础设施生命周期时才新增 Starter。
5. 云厂商相关能力统一放入对应厂商目录，不把厂商 SDK 泄漏进通用 Starter。
6. README 只提供索引和最小入口，详细配置留在各 Starter 文档中。
7. 依赖增加必须能对应一个已列入索引的能力或内置 driver。

单 Go Module 仍会在 `go.mod` 中记录全部直接依赖；这属于统一版本治理成本，不等于应用运行时加载全部依赖。若未来依赖下载体积成为有测量证据的主要问题，再单独评估多 Module，不提前增加发布复杂度。

## 4. Starter 索引

### 4.1 定位

仓库根目录增加 `starter-index.yaml`，作为 Nova Starter 与内置 driver 的发现事实源。README、组合矩阵和 AI Coding 指南引用该索引，但不从索引生成运行时代码。

### 4.2 最小结构

```yaml
schema_version: 1
starters:
  - id: novagorm
    package: github.com/luaxlou/nova/starter/gorm/novagorm
    category: data
    owner: nova
    status: stable
    config_root: gorm
    lifecycle: [lazy, named, reload, close]
    docs: docs/starters/novagorm.md
    capabilities:
      - gorm
      - model-first-guard
    drivers:
      - id: mysql
        status: stable
      - id: postgres
        status: stable
        verified_server: PostgreSQL 18
```

每个索引项必须包含稳定 ID、Go 包路径、能力分类、责任人、成熟度、配置根、生命周期、文档和能力列表。拥有内部 driver 的 Starter 还必须列出 driver ID、成熟度和已验证服务端版本。

### 4.3 治理规则

- 索引只声明仓库中真实存在、可构建、具有文档的能力。
- `experimental` 能力不得在 README 中描述为生产稳定。
- 删除或重命名稳定 ID、包路径、配置根属于兼容性变更。
- 新增 driver 是加法变更；不得改变现有 MySQL 默认行为。
- 索引校验只检查路径、文档和必要字段，不建立复杂代码生成链。

## 5. `novagorm` PostgreSQL 18 Driver

### 5.1 公共接口

现有公共调用保持不变：

```go
db, err := novagorm.DB()
analyticsDB, err := novagorm.Named("analytics").DB()
novagorm.Register("custom", builder)
err = novagorm.CloseAll()
```

新增 `OpenPostgresFromSQLDB(sqlDB *sql.DB) (*gorm.DB, error)`，与现有 `OpenMySQLFromSQLDB` 对称，并安装同一 Model First 门禁。除该显式桥接函数外，不增加 PostgreSQL 专属业务 API。

### 5.2 配置

单实例配置：

```yaml
gorm:
  driver: postgres
  postgres:
    dsn: "host=127.0.0.1 user=sbom password=<runtime-secret> dbname=sbom_platform port=5432 sslmode=disable TimeZone=UTC"
    max_open: 20
    max_idle: 10
    conn_max_lifetime: 1800
    conn_max_idle_time: 300
    prefer_simple_protocol: false
```

多实例继续使用既有结构：

```yaml
gorm:
  default: main
  main:
    driver: postgres
    postgres:
      dsn: "<runtime-secret>"
      max_open: 20
      max_idle: 10
  legacy:
    driver: mysql
    mysql:
      dsn: "<runtime-secret>"
```

规则：

- driver ID 固定为 `postgres`，不同时支持 `postgresql` 别名。
- PostgreSQL 配置只位于 `gorm.postgres` 或 `gorm.<name>.postgres`。
- `dsn` 必填；缺失时返回包含实例与配置路径的错误。
- 连接池数值小于等于零表示保留 Go SQL 默认值。
- DSN 和密码不得写入日志或错误信息。
- Starter 不硬编码服务端主版本；PostgreSQL 18 是本次真实集成验证基线。

### 5.3 内部实现边界

`gormConfig` 增加 `Postgres postgresConfig`。配置解析、driver 选择、连接创建和连接池设置继续保留在 `novagorm` 包中：

```text
novaconfig
  → parseGormConfig
  → newConfiguredConnection
  → newPostgresConnection
  → installAutoMigrateGuard
  → registry lifecycle
  → novagorm.DB()/Named().DB()
```

使用 GORM 官方 PostgreSQL dialector。`postgresConfig` 只保留建立连接和连接池所需的真实配置，不复制整个 PGX 配置面。`prefer_simple_protocol` 是首期唯一额外 driver 选项，用于需要关闭隐式 prepared statement 的环境。

### 5.4 Model First 一致性

PostgreSQL 与 MySQL 必须经过同一 `installAutoMigrateGuard`：

- 必须定义 `TableName()`。
- 所有持久化字段必须声明 `column`。
- 禁止 `gorm.Model`、`gorm.DeletedAt`、自动时间、生命周期 Hook 和 association。
- 应用明确列出 Model 后才可调用 `AutoMigrate`。
- Starter 不维护 SQL migration，也不自动调用 `AutoMigrate`。

门禁包装器必须保留 PostgreSQL Migrator 所需的扩展接口。如果 GORM PostgreSQL dialector 暴露了当前包装器未转发的接口，必须显式转发，而不是绕过门禁。

### 5.5 错误与生命周期

- 不支持的 driver：返回 `unsupported gorm driver`，不回退到 MySQL。
- 配置缺失：首次获取实例失败，不创建半初始化注册项。
- 连接创建或连接池访问失败：附加 driver/实例上下文，不包含 DSN。
- `Reload()`：关闭旧资源后按同一配置重新建立，不改变选中实例规则。
- `Close()` / `CloseAll()`：继续通过底层 `sql.DB.Close()` 释放连接。
- Starter 不负责业务错误翻译；唯一键、外键和序列化冲突由消费项目的数据能力在技术边界内转换。

## 6. 兼容性与依赖

- 新增 `gorm.io/driver/postgres` 直接依赖，并接受其 PGX 传递依赖。
- 现有 `gorm.driver: mysql`、`gorm.mysql.*` 和 `OpenMySQLFromSQLDB` 保持兼容。
- 未显式选择 `postgres` 的应用不改变连接行为。
- `Register` 继续允许真实外部变化点，不要求所有 driver 都进入 Nova。
- Nova 不提供 MySQL 到 PostgreSQL 的数据迁移工具；消费项目在尚无须保留历史业务数据时从空库初始化。

## 7. 未来验证要求

Nova 仓库验证分为三层：

1. 静态与单元验证：配置解析、保留键识别、driver 分派、缺失 DSN、连接池参数和索引完整性。
2. 门禁回归：MySQL 与 PostgreSQL 连接都安装同一 ORM Magic 门禁，现有 MySQL 行为不退化。
3. PostgreSQL 18 集成验证：使用临时空数据库建立真实连接，执行一组显式 GORM Model 初始化，验证表、主键、唯一索引、外键、JSONB、事务回滚和关闭行为。

Nova 完整 `go test ./...` 必须通过。真实 PostgreSQL 集成验证必须显式连接测试数据库，不能用 SQLite、Mock 或只验证 SQL 字符串替代。

## 8. 软件供应链安全整体计划中的位置

当整体设计冻结并明确进入相应建设阶段后，先交付 Nova PostgreSQL driver，再由软件供应链安全项目单独实施：

1. 锁定包含 PostgreSQL driver 的 Nova 版本。
2. 将最高工程约束、现行架构基线和运行配置统一为 PostgreSQL 18。
3. 移除 MySQL driver 直接依赖和 `mysql.MySQLError` 判断。
4. 依据现行 Model First 原则设计 SBOM、快照、组件、本体、证据和 HITL 持久化 Model。
5. 使用 PostgreSQL JSONB、部分索引和约束表达真实稳定需求；不把整个 LinkML 模型粗暴塞入单一 JSONB。
6. 将唯一黄金冒烟切换为空 PostgreSQL 18 数据库，再执行真实 `sbom submit`。

Nova driver 交付与消费项目迁移必须分成两个提交或 PR，避免基础设施能力与业务 Schema 相互污染。

## 9. 未来实施完成条件

- Nova 仍为单仓库、单 Go Module，没有新增外部 Starter 工程。
- Starter 索引覆盖仓库内全部公开 Starter，并能通过轻量校验。
- `novagorm` 在同一公共 API 下支持 MySQL 与 PostgreSQL。
- PostgreSQL 18 真实集成验证通过，MySQL 回归通过。
- 两种 driver 使用同一 Model First 门禁、多实例与生命周期语义。
- README、Starter 约定、组合矩阵、AI Coding 指南和 `novagorm` 专用文档口径一致。
- 软件供应链安全项目不再需要自行创建数据库生命周期或绕开 Nova。

## 10. 明确保留的后续事项

- 当前只把 PostgreSQL 18 支持列入整体计划，不从本文档直接进入编码或详细实施计划。
- 阿里云 Starter 的目录归并单独设计，不在 PostgreSQL 工作中移动 `novaqwen`。
- 只有真实依赖或下载成本数据证明单 Module 不可接受时，才重新评估多 Module。
- PostgreSQL 高可用、备份、PITR、连接代理和生产密钥管理属于部署设计，不进入通用 Starter。
