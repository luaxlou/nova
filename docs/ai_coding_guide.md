# Nova AI Coding 协作指南

本文档用于让 AI Agent 在 `nova` 仓库内快速建立上下文，并围绕统一 starter 设计哲学稳定交付。

开始具体编码前，必须同时阅读 `docs/nova_engineering_best_practices.md`，并把其中的业务域、动作文件、状态对象、Model First、data capability、adapter 边界作为默认工程判断。

## 一句话定位

`nova` 是应用侧框架仓库，提供 starter / sdk 能力，目标是让项目持续获得边界清晰、可组合、低心智负担和稳定约定。

## 引入后要拿到的收益

- 用统一初始化范式替代分散样板代码
- 通过 starter 组合减少重复封装
- 让团队在新项目与存量项目里使用一致接入语言
- 把接入沉淀为可演进工程基线

## 修改边界

允许改动：

- `starter/*`
- `examples/*`
- `docs/*`

禁止越界：

- 不实现与 starter / sdk 无关的系统能力
- 不引入与当前任务目标无关的额外职责
- 不做无关的大规模重构

## 标准执行顺序

1. 阅读流程卡：
   - 新项目：`docs/quickstart_new_project.md`
   - 存量项目：`docs/quickstart_existing_project.md`
2. 阅读工程最佳实践：`docs/nova_engineering_best_practices.md`。
3. 按业务域、业务动作、状态、数据模型、data capability、adapter 的顺序理解目标代码。
4. 按固定结构组织输出内容：实施计划、改动文件清单、收益说明、验证结果。
5. 执行统一验证命令并回报结果。

## 工程最佳实践约束

- 业务域优先：优先按 `internal/<domain>` 聚合业务上下文，而不是按 controller/service/repository 横向分层。
- 动作优先：一个业务动作一个文件，例如 `register.go` 对应 `user.Register(...)`。
- 状态决定实例：只有真实拥有状态、生命周期或多个实例的概念才建 struct。
- 能力直接表达：无状态能力优先使用 package function。
- Model First：数据库结构从 `data/model.go` 的 GORM model 出发，model 是当前数据结构的事实来源。
- Data 属于领域能力：`data/query.go`、`data/write.go`、`data/tx.go` 表达领域需要的数据能力，业务层不直接暴露 GORM、SQL、Redis 细节。
- HTTP 只是 adapter：request/response、status code、Gin context 留在 `http/` 包内，业务模型保持协议无关。
- Interface follows variation：只有真实变化点出现时才引入接口，并由使用方定义最小能力。

## AI 默认阅读顺序

处理业务需求时按这个顺序读取：

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

例如修改订单取消规则，优先读取 `internal/order/cancel.go`、`internal/order/order.go`、`internal/order/policy.go`；如果涉及数据，再读取 `internal/order/data/model.go` 和 `internal/order/data/query.go`。

## 统一验证命令

```bash
go test ./...
go vet ./...
```

## 快速索引

- `starter/config/novaconfig`
- `starter/http/novagin`
- `starter/gorm/novagorm`
- `starter/cache/novaredis`
- `starter/realtime/novawebsocket`
- `examples/simple-app`
- `examples/best-practice-service`
- `docs/nova_engineering_best_practices.md`
- `docs/starter_conventions.md`
- `docs/starter_composition_matrix.md`
