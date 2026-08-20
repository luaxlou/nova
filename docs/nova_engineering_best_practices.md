# Nova 工程最佳实践

本文档定义 Nova 推荐的应用工程方式。Nova 面向的是一种更贴近 Go、更适合 AI Coding 的工程模型：业务代码直接表达业务，package 直接表达领域和能力，function 直接表达动作，struct 承载真实状态，interface 表达真实变化点，Nova Starter 管理基础设施状态与生命周期，数据库采用 Model First，工程结构围绕业务组织。

核心思想：

> 业务就是业务。

> 状态决定实例。

> 能力直接表达。

> Model 定义数据。

> 技术复杂度停留在它真正所属的位置。

最终目标不是让代码看起来“有架构”，而是让代码本身就是架构。

## Nova 在应用中的位置

Nova 提供应用需要的基础设施能力：

```text
novaconfig
novagin
novamysql
novagorm
novaredis
novawebsocket
```

它们负责基础设施本身的状态和生命周期。

```text
Application
    │
    ├── novaconfig
    │       └── Configuration
    ├── novagin
    │       └── HTTP
    ├── novagorm
    │       └── GORM
    │           ↓
    ├── novamysql
    │       └── MySQL / *sql.DB
    ├── novaredis
    │       └── Redis
    └── novawebsocket
            └── WebSocket
```

数据库能力进一步分成：

```text
novagorm
    ↓
novamysql
```

`novamysql` 管理数据库连接资源：`*sql.DB`、连接池、命名实例和生命周期。`novagorm` 在其上提供 ORM 能力：`*gorm.DB`、GORM adapter 和命名实例。应用只使用自己需要的能力：

```go
db, err := novamysql.DB()
```

或者：

```go
db, err := novagorm.DB()
```

多实例：

```go
db, err := novagorm.Named("analytics").DB()
```

状态属于基础设施，因此由基础设施自己管理。

## Nova 应用的整体模型

Nova 应用推荐形成：

```text
Entry
  ↓
Adapter
  ↓
Business
  ↓
Capability
  ↓
Nova Starter
  ↓
Infrastructure
```

例如：

```text
HTTP Request
    ↓
user/http.Register
    ↓
user.Register
    ↓
user/data.EmailExists
    ↓
novagorm.DB
    ↓
novamysql.DB
    ↓
MySQL
```

每一层只表达自己的责任：

- Adapter：外部世界如何进入业务。
- Business：系统正在完成什么业务。
- Capability：业务需要什么能力。
- Nova Starter：基础设施能力如何获得。
- Infrastructure：最终技术如何运行。

## 推荐项目目录

一个典型 Nova 业务项目：

```text
.
├── cmd/
│   ├── api/
│   │   └── main.go
│   └── worker/
│       └── main.go
├── internal/
│   ├── user/
│   │   ├── user.go
│   │   ├── register.go
│   │   ├── login.go
│   │   ├── policy.go
│   │   ├── errors.go
│   │   ├── data/
│   │   │   ├── model.go
│   │   │   ├── query.go
│   │   │   ├── write.go
│   │   │   └── tx.go
│   │   └── http/
│   │       ├── routes.go
│   │       ├── register.go
│   │       ├── request.go
│   │       └── response.go
│   ├── order/
│   │   ├── order.go
│   │   ├── create.go
│   │   ├── cancel.go
│   │   ├── policy.go
│   │   ├── errors.go
│   │   ├── data/
│   │   └── http/
│   ├── integration/
│   │   ├── mail/
│   │   ├── sms/
│   │   └── eventbus/
│   ├── httpx/
│   ├── shared/
│   └── tool/
├── test/
│   ├── e2e/
│   └── fixture/
├── scripts/
├── docs/
├── config.yaml
├── go.mod
└── go.sum
```

核心不是目录数量，而是先按业务域组织，再在业务域内部表达不同能力。项目第一眼看到的是 `user`、`order`、`payment`，因此工程结构首先回答：这个系统有什么业务。

## cmd：应用入口

`cmd` 表示真正可以运行的程序。不同入口共同使用同一套业务能力。

`cmd/api/main.go` 负责应用启动、starter 初始化、model 初始化、路由组合、server 启动和生命周期管理。

```go
func main() {
    if err := novagorm.Init(); err != nil {
        log.Fatal(err)
    }

    if err := initModels(); err != nil {
        log.Fatal(err)
    }

    if err := novaredis.Init(); err != nil {
        log.Fatal(err)
    }

    r := novagin.Router()
    userhttp.Routes(r)
    orderhttp.Routes(r)
    paymenthttp.Routes(r)

    novagin.Init(8080)
    novagin.Run()

    select {}
}
```

应用入口负责把运行系统组合起来，业务实现仍然存在于业务域。

## 业务域是工程主体

Nova 项目的主体是 `internal/user`、`internal/order`、`internal/payment` 这样的真实业务概念。修改用户注册时进入 `internal/user`，修改订单取消时进入 `internal/order`，一个业务问题所需要的上下文自然聚集在同一个领域。

## 一个业务动作一个文件

业务流程按照业务动作组织：

```text
user/
├── register.go
├── login.go
├── change_password.go
└── disable.go
```

文件名、package 和 function 形成一致语义：

```text
order/cancel.go
    ↓
order.Cancel(...)
```

```go
func Cancel(ctx context.Context, orderID string) error {
    record, err := data.Find(ctx, orderID)
    if err != nil {
        return err
    }

    o := restore(record)

    if err := o.Cancel(); err != nil {
        return err
    }

    return data.UpdateStatus(ctx, o.ID, string(o.Status))
}
```

代码直接表达：找到订单，恢复订单，执行取消，保存结果。

## Package 表达能力

Nova 把 package 作为主要的能力边界：

```go
user.Register(...)
user.Login(...)

order.Create(...)
order.Cancel(...)

payment.Pay(...)
payment.Refund(...)
```

`user`、`order`、`payment` 表达领域，`Register`、`Cancel`、`Pay` 表达动作。Package 自身已经承担了足够清晰的语义边界。

## 状态决定实例

Struct 用于表达真正存在的状态：

```go
type Order struct {
    ID     string
    Status Status
    Amount money.Money
}
```

Order 拥有身份、状态、生命周期和业务约束，因此自然拥有：

```go
func (o *Order) Cancel() error
func (o *Order) Pay() error
func (o *Order) Complete() error
```

无状态行为直接使用 package function：

```go
user.Register(...)
mail.Send(...)
crypto.Hash(...)
id.New(...)
clock.Now(...)
```

Nova 的基本映射关系：

```text
Package  = 能力
Function = 动作
Struct   = 状态
```

## 业务就是业务

业务代码最重要的目标是：打开代码，看到的是业务过程。

```go
func Register(ctx context.Context, cmd RegisterCommand) (*User, error) {
    if err := validateRegister(cmd); err != nil {
        return nil, err
    }

    exists, err := data.EmailExists(ctx, cmd.Email)
    if err != nil {
        return nil, err
    }

    if exists {
        return nil, ErrEmailExists
    }

    u := New(cmd)

    if err := data.Insert(ctx, data.InsertParams{
        ID:    u.ID,
        Email: u.Email,
        Name:  u.Name,
    }); err != nil {
        return nil, err
    }

    return u, nil
}
```

阅读代码得到：验证注册信息，判断邮箱是否存在，创建用户，保存用户。业务层回答的是系统正在做什么。

## Policy 表达业务规则

纯业务规则直接表达：

```go
func CanCancel(status Status) bool {
    return status == StatusCreated || status == StatusPaid
}
```

规则较多后可以按照真实业务概念拆分为 `cancel_policy.go`、`refund_policy.go`、`price_policy.go`。

## 领域错误

领域错误和业务放在一起：

```go
var (
    ErrOrderNotFound = errors.New("order not found")
    ErrCannotCancel  = errors.New("order cannot be cancelled")
    ErrAlreadyPaid   = errors.New("order already paid")
)
```

HTTP adapter 再决定它对应 `400`、`404` 或 `409`。业务保持协议无关。

## Data 是领域的数据能力

每个业务域可以拥有 `data/`：

```text
user/data/
├── model.go
├── query.go
├── write.go
└── tx.go
```

它表达 User 领域需要哪些数据能力，例如 `data.EmailExists(...)`、`data.Find(...)`、`data.Insert(...)`、`data.UpdateProfile(...)`。业务仍然看到业务语义，GORM、Redis、MySQL 属于 data 内部实现。

## Model First

Nova 的数据库工程采用 Model First。所有数据库设计都首先从 Model 出发。

```go
type UserModel struct {
    ID string `gorm:"primaryKey;size:32"`

    Email string `gorm:"size:255;uniqueIndex;not null"`
    Name  string `gorm:"size:100;not null"`

    Status string `gorm:"size:32;index;not null"`

    CreatedAt time.Time
    UpdatedAt time.Time
}
```

这个 Model 同时表达数据结构、字段类型、长度、主键、索引、唯一约束、关联和数据库约束。因此数据库设计的起点始终是 Model。

```text
Model
    ↓
GORM
    ↓
Database
```

Model 是数据层的事实来源。当前系统的数据定义应该能够直接从代码中的 Model 得到。业务模型和 data model 分别回答两个问题：

```text
Order      -> 业务中的订单是什么？
OrderModel -> 订单如何持久化？
```

业务模型保持纯净，data model 完整承担数据库表达。

新环境启动数据库时，根据当前 Model 建立数据库结构：

```go
func initModels() error {
    db, err := novagorm.DB()
    if err != nil {
        return err
    }

    return db.AutoMigrate(
        &userdata.UserModel{},
        &orderdata.OrderModel{},
        &paymentdata.PaymentModel{},
    )
}
```

真正重要的不是 `AutoMigrate` 本身，而是数据库结构来源等于当前 Model。

数据结构演进仍然从 Model 开始：

```text
业务需要新的数据
↓
修改 Model
↓
代码以新 Model 为准
↓
数据库适应 Model
```

有些变化不仅仅是结构变化，例如历史数据转换、字段语义变化、数据合并、数据拆分、大规模数据重算。这种情况本质上是数据维护任务，可以编写一次性程序放在 `cmd/maintenance/...`，长期的数据结构定义仍然回归 `data/model.go`。

## Model First 对 AI Coding 的意义

当 AI 想理解用户表时，只需要读取 `user/data/model.go` 即可理解字段、类型、索引、关系和约束。当 AI 修改业务数据结构时，也可以从业务需求到 Model，再到 Data Capability，直接完成修改。

因此当前代码就是当前系统，这能显著降低 AI 理解系统所需要的上下文。

## Query 与 Write 表达数据能力

查询能力：

```go
func EmailExists(ctx context.Context, email string) (bool, error) {
    db, err := novagorm.DB()
    if err != nil {
        return false, err
    }

    var count int64

    err = db.WithContext(ctx).
        Model(&UserModel{}).
        Where("email = ?", email).
        Count(&count).
        Error

    return count > 0, err
}
```

写入能力：

```go
type InsertParams struct {
    ID    string
    Email string
    Name  string
}

func Insert(ctx context.Context, p InsertParams) error {
    db, err := novagorm.DB()
    if err != nil {
        return err
    }

    return db.WithContext(ctx).
        Create(&UserModel{
            ID:    p.ID,
            Email: p.Email,
            Name:  p.Name,
        }).
        Error
}
```

业务调用：

```go
data.Insert(ctx, data.InsertParams{
    ID:    u.ID,
    Email: u.Email,
    Name:  u.Name,
})
```

依赖方向保持：

```text
user
 ↓
user/data
 ↓
novagorm
 ↓
novamysql
```

Data Capability 可以定义自己的边界类型，例如 `InsertParams`、`UpdateParams`、`QueryResult`、`UserModel`，这样 business 到 data 始终保持单向依赖。

## Transaction 与 Cache

事务拥有局部状态和生命周期，因此可以成为实例：

```go
err := data.WithTx(ctx, func(tx *data.Tx) error {
    if err := tx.InsertOrder(...); err != nil {
        return err
    }

    return tx.InsertPayment(...)
})
```

缓存通常是 Data Capability 的实现细节：

```text
user/data/
├── model.go
├── query.go
├── write.go
└── cache.go
```

业务仍然调用 `data.Find(ctx, id)`，内部可以 Redis miss 后查询 MySQL，再写回 Redis。

## HTTP 是协议 Adapter

每个业务域可以拥有 `http/`：

```text
user/http/
├── routes.go
├── register.go
├── login.go
├── request.go
└── response.go
```

HTTP 负责把协议转换成业务调用。业务模型不会携带 JSON tag、HTTP status、HTTP header 或 Gin context。

Routes 负责路由：

```go
func Routes(r *gin.Engine) {
    group := r.Group("/users")

    group.POST("", Register)
    group.POST("/login", Login)
    group.GET("/:id", Profile)
}
```

Handler 只负责协议转换：

```go
func Register(c *gin.Context) {
    var req RegisterRequest

    if err := c.ShouldBindJSON(&req); err != nil {
        writeBadRequest(c, err)
        return
    }

    result, err := user.Register(
        c.Request.Context(),
        user.RegisterCommand{
            Email: req.Email,
            Name:  req.Name,
        },
    )

    writeRegisterResult(c, result, err)
}
```

过程是 HTTP Request 到 Business Command，到 Business，再到 HTTP Response。

跨多个业务域的 HTTP 能力放在 `internal/httpx/`，例如 authentication adapter、trace、recovery、统一 response 和 middleware。

## Integration、Shared 与 Tool

外部系统能力统一放在 `internal/integration/`：

```text
integration/
├── mail/
├── sms/
├── paymentgateway/
└── eventbus/
```

业务看到的是能力语义，例如 `mail.SendVerification(...)`、`eventbus.Publish(...)`。无状态 integration 直接通过 function 表达，拥有独立状态、生命周期、多个实例或运行时选择时，再自然形成对象。

多个领域共同使用，并且具有明确业务意义的概念放在 `shared/`，例如 money、identity、paging。纯工具放在 `tool/`，例如 crypto、id、clock、text。判断标准只有一个：它是不是业务语言的一部分。

## Interface 表达真实变化点

Interface follows variation。变化出现在哪里，抽象就出现在哪里。

```go
type Gateway interface {
    Pay(ctx context.Context, req Request) (*Result, error)
}
```

接口由使用方定义。如果调用方只需要 `Find(ctx, id)`，接口就可以只有：

```go
type UserFinder interface {
    Find(ctx context.Context, id string) (*User, error)
}
```

接口表达“我需要什么能力”，而不是描述实现方拥有多少功能。

## 跨领域协作与 Event

业务域通过业务能力互相协作：

```go
result, err := payment.Create(ctx, payment.CreateCommand{
    OrderID: o.ID,
    Amount:  o.Amount,
})
```

Order 调用的是 Payment Business Capability，而不是 Payment 的数据库实现。

Event 首先是业务概念：

```go
type OrderCreated struct {
    OrderID string
    UserID  string
}
```

Kafka、RabbitMQ、Redis Stream 等属于 integration 实现。

## 测试跟随工程边界

普通 Go test 和实现代码放在一起：

```text
order/
├── order.go
├── order_test.go
├── cancel.go
├── cancel_test.go
└── data/
    ├── query.go
    └── query_test.go
```

业务规则测试直接验证业务规则，data capability 测试验证 GORM model、query、write、constraint、relation 和 transaction。跨领域测试放在 `test/e2e/`。

## 多数据库

Nova 支持 named instance：

```text
main
analytics
archive
```

Data Capability 内：

```go
db, err := novagorm.Named("analytics").DB()
```

业务则可以调用：

```go
analytics.UserActivity(...)
```

数据库选择停留在真正属于它的数据能力中。

## 工程从最小结构开始

新项目完全可以只有：

```text
cmd/
└── api/
    └── main.go

internal/
└── user/
    ├── register.go
    ├── data/
    │   └── model.go
    └── http/
        └── register.go
```

业务增长之后再自然成长。

新增业务时：

```text
业务概念
↓
业务动作
↓
领域状态
↓
业务规则
↓
Model
↓
数据能力
↓
外部能力
↓
Adapter
```

所有结构从真实需求向外自然生长。

## AI Coding 的默认理解顺序

AI Agent 进入 Nova 项目时，默认按这个顺序理解：

```text
目标业务域
↓
目标业务动作
↓
Domain Model / Policy
↓
Data Model
↓
Data / Integration
↓
Adapter
↓
Nova Starter
```

例如修改订单取消规则，优先读取：

```text
internal/order/cancel.go
internal/order/order.go
internal/order/policy.go
```

如果涉及数据，再进入：

```text
internal/order/data/model.go
internal/order/data/query.go
```

最后才需要理解底层基础设施。

## AI Coding 的工程判断

新增代码时依次判断：

- 这是什么业务概念？决定 package。
- 它正在完成什么动作？决定 function。
- 它是否拥有状态？决定 struct。
- 状态属于哪里？决定生命周期。
- 它的数据是什么？决定 model。
- 它需要什么数据能力？决定 data。
- 它需要什么外部能力？决定 integration。
- 是否存在真实变化？决定 interface。

最后再确认：代码是否仍然直接表达真实系统。

## Nova 的工程映射

```text
业务领域       = Package
业务动作       = Function
业务状态       = Struct
业务规则       = Pure Function / Domain Method
持久化模型     = GORM Model
数据设计       = Model First
数据能力       = data
HTTP          = http Adapter
共享 HTTP 能力 = httpx
外部系统       = integration
共享业务概念   = shared
通用工具       = tool
真实变化点     = Interface
进程级基础设施 = Nova Starter
```

一个成熟 Nova 项目应该自然出现：

```go
user.Register(...)
user.Login(...)
order.Create(...)
order.Cancel(...)
payment.Pay(...)
mail.Send(...)
data.Find(...)
data.Insert(...)
novagorm.DB()
novaredis.Client()
```

每一个名字都对应真实含义。

最终 Nova 形成的是：

> 业务直接表达业务。

> Package 直接表达能力。

> Function 直接表达动作。

> Struct 承载真实状态。

> Model First 定义数据。

> Starter 管理基础设施。

> 工程结构与真实系统保持一致。

这样，人和 AI 都可以用最小上下文理解系统，并让系统随着业务自然生长。
