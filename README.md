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
目标：基于 nova 建立统一、可组合的应用接入基线，并快速落地最小可运行服务。
重点收益：通过 nova 的稳定约定、可组合能力和低心智负担，提升项目可维护性与团队交付效率。
工程要求：按业务域组织代码；业务动作使用 package function 直接表达；有状态对象使用 struct；数据设计采用 Model First；HTTP 仅作为 adapter。
需解决问题：如何让项目在引入后持续享受 nova 设计哲学带来的收益，而不是一次性接入。
输出要求：给出实施计划、文件改动清单、收益说明与验证结果。
完成后执行并反馈：go test ./... && go vet ./...
```

### 模板 2：现有项目引入 nova

```text
请在现有项目中引入 nova starter。
请先阅读：https://github.com/luaxlou/nova/blob/main/docs/ai_coding_guide.md
并阅读 Nova 工程最佳实践：https://github.com/luaxlou/nova/blob/main/docs/nova_engineering_best_practices.md
目标：建立 nova 的统一接入范式，减少重复样板并提升长期可维护性。
重点收益：让项目获得边界清晰、可组合复用、低心智负担和稳定约定。
工程要求：识别现有业务域、业务动作、数据模型与 adapter 边界；优先让业务代码直接表达业务；数据库改动从 Model 开始。
需解决问题：如何把接入过程沉淀为可持续演进的工程基线。
输出要求：给出实施步骤、受影响文件清单、收益说明与验证结果。
完成后执行并反馈：go test ./... && go vet ./...
```

## 这个仓库包含什么

- [`starter/config/novaconfig`](./starter/config/novaconfig)：配置读取；说明见 [`docs/starters/novaconfig.md`](./docs/starters/novaconfig.md)
- [`starter/http/novagin`](./starter/http/novagin)：HTTP 服务启动适配（Gin）；说明见 [`docs/starters/novagin.md`](./docs/starters/novagin.md)
- [`starter/cache/novaredis`](./starter/cache/novaredis)：Redis 客户端初始化；说明见 [`docs/starters/novaredis.md`](./docs/starters/novaredis.md)
- [`starter/aliyun/novaoss`](./starter/aliyun/novaoss)：Alibaba Cloud OSS Bucket 初始化；说明见 [`docs/starters/novaoss.md`](./docs/starters/novaoss.md)
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
