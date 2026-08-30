# Nova

## 诞生理念

`nova` 的目标不是增加一个大而全框架，而是把 Go 项目里的高频基础接入沉淀成稳定约定，让团队持续获得这些收益：

- 边界清晰：业务逻辑与基础接入职责分离
- 可组合：按需引入 starter，避免绑定整套技术栈
- 低心智负担：统一初始化范式，降低协作与接手成本
- 稳定约定：减少重复样板，提升长期可维护性

## AI Coding 快速引入（可直接复制）

AI 协作文档（完整链接）：
https://github.com/luaxlou/nova/blob/main/docs/ai_coding_guide.md

Nova 工程最佳实践（完整链接）：
https://github.com/luaxlou/nova/blob/main/docs/nova_engineering_best_practices.md

新项目流程卡（5 步）：
https://github.com/luaxlou/nova/blob/main/docs/quickstart_new_project.md

存量项目流程卡（5 步）：
https://github.com/luaxlou/nova/blob/main/docs/quickstart_existing_project.md

### 提示词模版：新项目引入 nova

```text
请基于 nova 初始化一个新服务。
请先阅读：https://github.com/luaxlou/nova/blob/main/docs/ai_coding_guide.md
并阅读 Nova 工程最佳实践：https://github.com/luaxlou/nova/blob/main/docs/nova_engineering_best_practices.md
引入门禁：既然选择引入 nova，就必须完整遵循 Nova 工程最佳实践，不允许只接入 starter 而保留与最佳实践冲突的工程方式。如果无法或不愿完整遵循，请不要引入 nova。
目标：基于 nova 建立统一、可组合的应用接入基线，并快速落地最小可运行服务。
重点收益：通过 nova 的稳定约定、可组合能力和低心智负担，提升项目可维护性与团队交付效率。
工程要求：按业务域组织代码；业务动作使用 package function 直接表达；有状态对象使用 struct；数据设计采用 Model First；HTTP 仅作为 adapter。
需解决问题：如何让项目在引入后持续享受 nova 设计哲学带来的收益，而不是一次性接入。
实施前检查：识别现有设计与 Nova 工程最佳实践的冲突；能够解决时把修正纳入实施计划，无法解决时停止引入并说明原因。
输出要求：给出最佳实践符合性检查、实施计划、文件改动清单、收益说明与现实验证结果。
完成后运行真实系统、走受影响的真实用户路径并反馈日志、指标或数据证据；go test ./... 与 go vet ./... 仅作为辅助检查。
```

### 模板 2：现有项目引入 nova

```text
请在现有项目中引入 nova starter。
请先阅读：https://github.com/luaxlou/nova/blob/main/docs/ai_coding_guide.md
并阅读 Nova 工程最佳实践：https://github.com/luaxlou/nova/blob/main/docs/nova_engineering_best_practices.md
引入门禁：既然选择引入 nova，就必须完整遵循 Nova 工程最佳实践，不允许只接入 starter 而保留与最佳实践冲突的工程方式。如果无法或不愿完整遵循，请不要引入 nova。
目标：建立 nova 的统一接入范式，减少重复样板并提升长期可维护性。
重点收益：让项目获得边界清晰、可组合复用、低心智负担和稳定约定。
工程要求：识别现有业务域、业务动作、数据模型与 adapter 边界；优先让业务代码直接表达业务；数据库改动从 Model 开始。
需解决问题：如何把接入过程沉淀为可持续演进的工程基线。
实施前检查：识别现有设计与 Nova 工程最佳实践的冲突；能够解决时把修正纳入实施计划，无法解决时停止引入并说明原因。
输出要求：给出最佳实践符合性检查、实施步骤、受影响文件清单、收益说明与现实验证结果。
完成后运行真实系统、走受影响的真实用户路径并反馈日志、指标或数据证据；go test ./... 与 go vet ./... 仅作为辅助检查。
```

### 提示词模板：Golang 项目 Nova 最佳实践审查

```text
请使用 Nova 工程最佳实践审查当前整个 Golang 项目。本审查适用于任何 Golang 项目，不要求项目已经引入 nova，也不要求为了通过审查而引入任何 Nova Starter。

请先完整阅读 Nova 工程最佳实践：
https://github.com/luaxlou/nova/blob/main/docs/nova_engineering_best_practices.md

本次只审查并制定调整计划，不要直接修改代码。审查必须基于当前项目的真实目录和源码证据，不要根据命名或惯例臆测。

重点检查：
1. 目录结构与代码归属：是否按真实业务域组织；业务动作、状态对象、data capability、HTTP adapter、integration、shared 和 tool 是否放在正确位置；文件内容是否与目录职责一致；是否存在 controller/service/repository 等按技术职责横向切割业务的问题。
2. 无状态优先：没有真实状态、生命周期或多实例需求的能力，是否直接使用 package function；是否存在仅用于承载方法的 Service、Manager、Handler 等无状态 struct。
3. 反 DI：是否存在 DI 容器、Provider/Wire 图、层层构造器注入、仅为注入依赖而存在的对象，或仅为 Mock 预设的接口。真实拥有状态、生命周期、多实例或运行时选择的对象不属于问题；已经出现真实变化点时，可以由使用方定义最小接口。
4. Model First：如果项目存在数据库持久化，检查 Model 是否是数据库结构的唯一事实来源，表、字段、索引和约束是否由 Model 明确表达，业务模型、持久化 Model 与 data capability 的边界是否清晰，以及是否存在重复维护的建表 SQL、迁移定义或隐式 ORM 行为。如果项目没有数据库，明确标记为“不适用”。

审查整个项目，但排除 vendor、第三方依赖、生成代码、构建产物和缓存目录。不要把未使用的数据库、ORM、AI、缓存或其他 Starter 当成必选项。

请直接输出：
1. 整体符合性结论，以及结论所依据的代码范围。
2. 当前关键目录树与建议目标目录树。
3. 按影响排序的偏离项；每项必须包含文件路径和代码位置、违反的原则、现实影响与调整建议。没有证据的问题不要列出。
4. 可执行的调整计划；按阶段列出目标、涉及文件、具体动作、依赖关系和完成标准。
5. Reality-Driven Development 验证计划；说明调整后要运行什么、走哪些真实用户路径、观察哪些日志、指标或数据。go test ./... 与 go vet ./... 只能作为辅助检查。

输出计划后停止，不要执行任何修改。最后明确询问：是否按以上计划执行调整？只有得到确认后才能开始修改代码。
```

## 这个仓库包含什么

- [`starter/config/novaconfig`](./starter/config/novaconfig)：配置读取；说明见 [`docs/starters/novaconfig.md`](./docs/starters/novaconfig.md)
- [`starter/http/novagin`](./starter/http/novagin)：HTTP 服务启动适配（Gin）；说明见 [`docs/starters/novagin.md`](./docs/starters/novagin.md)
- [`starter/cache/novaredis`](./starter/cache/novaredis)：Redis 客户端初始化；说明见 [`docs/starters/novaredis.md`](./docs/starters/novaredis.md)
- [`starter/aliyun/novaoss`](./starter/aliyun/novaoss)：Alibaba Cloud OSS Bucket 初始化；说明见 [`docs/starters/novaoss.md`](./docs/starters/novaoss.md)
- [`starter/ai/novaqwen`](./starter/ai/novaqwen)：Alibaba Cloud Qwen Client 与 Chat Completions；说明见 [`docs/starters/novaqwen.md`](./docs/starters/novaqwen.md)
- [`starter/realtime/novawebsocket`](./starter/realtime/novawebsocket)：WebSocket 适配；说明见 [`docs/starters/novawebsocket.md`](./docs/starters/novawebsocket.md)
- [`starter/gorm/novagorm`](./starter/gorm/novagorm)：GORM Starter；说明见 [`docs/starters/novagorm.md`](./docs/starters/novagorm.md)
- [`examples/`](./examples)：可运行示例集合
- [`docs/sdk_manual.md`](./docs/sdk_manual.md)：SDK 手册
- [`docs/starter_conventions.md`](./docs/starter_conventions.md)：Starter 统一约定
- [`docs/starter_composition_matrix.md`](./docs/starter_composition_matrix.md)：Starter 组合矩阵
- [`docs/nova_engineering_best_practices.md`](./docs/nova_engineering_best_practices.md)：Nova 工程最佳实践

当前二代默认配置文件为 `config.yaml`（YAML）。需要读取配置的 starter 统一通过 `novaconfig` 获取配置；完整配置约定见 [`docs/starter_conventions.md`](./docs/starter_conventions.md) 与各 starter 专用说明。

最小配置示例：

```yaml
# starter/http/novagin
http:
  port: 8080

# starter/gorm/novagorm
gorm:
  main:
    driver: mysql
    mysql:
      dsn: root:password@tcp(localhost:3306)/app?parseTime=true
      max_open: 20
      max_idle: 10
  analytics:
    driver: mysql
    mysql:
      dsn: analytics:password@tcp(localhost:3306)/analytics?parseTime=true

# starter/cache/novaredis
redis:
  addr: localhost:6379
  db: 0

# starter/aliyun/novaoss
aliyun:
  oss:
    endpoint: https://oss-<region>.aliyuncs.com
    bucket: <bucket>
    access_key_id: <runtime-secret>
    access_key_secret: <runtime-secret>
```

## GORM / MySQL 约定

MySQL 不再作为独立 Starter 对外提供，只是 [`starter/gorm/novagorm`](./starter/gorm/novagorm) 的一种 driver 选择。GORM 支持多实例，实例直接放在 `gorm.<name>` 下：通过 `driver` 选择数据库类型，再把对应数据库配置放到 `mysql` 等 driver 节点下。只有一个实例时可以使用 `novagorm.DB()`；有多个实例时必须使用 `novagorm.Named("<name>").DB()`。

Nova 项目禁用 ORM Magic：表列映射、时间字段、外键和更新列都必须显式表达；Schema 坚持 Model First，保留基于 Model 的显式 `AutoMigrate`，并由 `novagorm` 在任何 DDL 前执行强制 Model 门禁。完整约束见 [`novagorm` 说明](./docs/starters/novagorm.md#禁用-orm-magic)。

## Starter 专用说明

每个 starter 都应有一份专用说明，用来回答三个问题：

- 配置从哪里读，使用哪些 key
- 最小接入代码是什么
- 与其他 starter 的组合边界是什么

当前说明入口：

- [`novaconfig`](./docs/starters/novaconfig.md)
- [`novagin`](./docs/starters/novagin.md)
- [`novagorm`](./docs/starters/novagorm.md)
- [`novaredis`](./docs/starters/novaredis.md)
- [`novaoss`](./docs/starters/novaoss.md)
- [`novaqwen`](./docs/starters/novaqwen.md)
- [`novawebsocket`](./docs/starters/novawebsocket.md)

## Alibaba Cloud OSS 约定

`starter/aliyun/novaoss` 使用 Alibaba Cloud OSS 官方 Go SDK，按 `Get/Named + Bucket + Reload/Close` 方式提供对象存储 Bucket。配置从 `aliyun.oss` 读取；访问密钥和临时令牌必须来自运行时配置或密钥管理系统，禁止写入源代码或提交到仓库。

```yaml
aliyun:
  oss:
    endpoint: https://oss-<region>.aliyuncs.com
    bucket: <bucket>
    access_key_id: <runtime-secret>
    access_key_secret: <runtime-secret>
    security_token: <optional-runtime-secret>
```

```go
import "github.com/luaxlou/nova/starter/aliyun/novaoss"

bucket, err := novaoss.Bucket()
```

多个 Bucket 时，直接配置在 `aliyun.oss.<name>` 下；通过 `novaoss.Named("<name>").Bucket()` 获取指定 Bucket。只有一个 Bucket 时可以直接使用 `novaoss.Bucket()`。

## 快速开始

```bash
go get github.com/luaxlou/nova
```

示例：

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

## 开发与验证

```bash
go test ./...
go vet ./...
```
