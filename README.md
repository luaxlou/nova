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

- [`starter/novaconfig`](./starter/novaconfig)：配置读取
- [`starter/novagin`](./starter/novagin)：HTTP 服务启动适配（Gin）
- Nova MySQL [`starter/novamysql`](./starter/novamysql)：MySQL 初始化
- Nova GORM [`starter/novagorm`](./starter/novagorm)：GORM 动态装配，支持独立 DSN 或自定义 builder
- [`starter/novaredis`](./starter/novaredis)：Redis 初始化
- Alibaba Cloud OSS [`starter/aliyun/oss`](./starter/aliyun/oss)：对象存储 Bucket 初始化
- [`starter/novawebsocket`](./starter/novawebsocket)：WebSocket 适配
- [`examples/`](./examples)：示例集合
- [`docs/sdk_manual.md`](./docs/sdk_manual.md)：SDK 手册
- [`docs/nova_engineering_best_practices.md`](./docs/nova_engineering_best_practices.md)：Nova 工程最佳实践

新增约定：当前二代默认配置文件为 `config.yaml`（YAML）。

## Nova MySQL / Nova GORM 约定

`starter/novamysql` 只负责 MySQL 原生连接初始化，向上提供 `*sql.DB`，不依赖 ORM。

需要 GORM 时引入 `starter/novagorm`。Nova GORM 是 GORM 的统一装配层，不依赖 Nova MySQL：可以通过 `gorm.dsn` 独立初始化，也可以通过 `Register` 注入自定义 builder。需要复用 Nova MySQL 时，由应用层把 `novamysql.DB()` 通过 `Register` 动态装配进来。

## Alibaba Cloud OSS 约定

`starter/aliyun/oss` 使用 Alibaba Cloud OSS 官方 Go SDK，按 `Init + Get/Named + Bucket + Reload/Close` 方式提供对象存储 Bucket。配置只从 `novaconfig` 读取；访问密钥和临时令牌必须来自运行时配置或密钥管理系统，禁止写入源代码或提交到仓库。

```yaml
oss:
  endpoint: https://oss-<region>.aliyuncs.com
  bucket: <bucket>
  access_key_id: <runtime-secret>
  access_key_secret: <runtime-secret>
  security_token: <optional-runtime-secret>
```

```go
import oss "github.com/luaxlou/nova/starter/aliyun/oss"

bucket, err := oss.Bucket()
```

多个 Bucket 时，设置 `oss.default` 并将各配置放在 `oss.instances.<name>` 下；通过 `oss.Named("<name>").Bucket()` 获取指定 Bucket。

## 快速开始

```bash
go get github.com/luaxlou/nova/starter
```

示例：

```go
package main

import (
    "fmt"

    "github.com/luaxlou/nova/starter/novaconfig"
)

func main() {
    fmt.Println(novaconfig.GetString("log_level"))
}
```

## 开发与验证

```bash
go test ./...
go vet ./...
```
