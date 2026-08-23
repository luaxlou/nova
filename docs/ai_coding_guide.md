# Nova AI Coding 协作指南

本文档用于让 AI Agent 在 `nova` 仓库内快速建立上下文，并围绕统一 starter 设计哲学稳定交付。

开始具体编码前，必须同时阅读 `docs/nova_engineering_best_practices.md`，并把其中的全部约束作为工程判断。Nova 不是可以脱离其工程方式单独接入的工具集合：既然引入 Nova，就必须完整遵循 Nova 工程最佳实践；如果无法或不愿完整遵循，就不要引入 Nova。

## 引入门禁

AI Agent 在新项目或存量项目中引入 Nova 前，必须先检查目标工程是否能够完整遵循 Nova 工程最佳实践：

- 不允许只接入 Nova Starter，同时保留与最佳实践冲突的分层、数据模型、隐式 ORM 行为或验证方式。
- 发现冲突时，必须明确列出冲突，并把修正纳入同一次引入计划。
- 如果关键冲突无法在当前范围内解决，必须停止引入 Nova，说明阻塞原因和所需条件，不能交付一个部分遵循的 Nova 项目。
- 只有完成符合性检查并确认能够完整遵循后，才可以开始接入。

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
3. 执行最佳实践符合性检查；存在冲突时将修正纳入计划，关键冲突无法解决时停止引入。
4. 按业务域、业务动作、状态、数据模型、data capability、adapter 的顺序理解目标代码。
5. 按固定结构组织输出内容：符合性检查、实施计划、改动文件清单、收益说明、现实验证结果。
6. 运行真实系统，走受影响的真实用户路径，并回报日志、指标或数据证据；测试和静态检查只作为辅助验证。

## 工程最佳实践约束

- 业务域优先：优先按 `internal/<domain>` 聚合业务上下文，而不是按 controller/service/repository 横向分层。
- 动作优先：一个业务动作一个文件，例如 `register.go` 对应 `user.Register(...)`。
- 状态决定实例：只有真实拥有状态、生命周期或多个实例的概念才建 struct。
- 能力直接表达：无状态能力优先使用 package function。
- Model First：数据库结构从 `data/model.go` 的 GORM model 出发，model 是当前数据结构的事实来源。
- 禁用 ORM Magic：显式声明表名、列名、时间字段、更新列和外键 ID，不声明 GORM association，禁止 hook、`gorm.Model` 和 `gorm.DeletedAt`；保留基于 Model 的显式 `AutoMigrate`，由 `novagorm` 在任何 DDL 前执行 Model 门禁，不维护重复的建表 SQL。
- Data 属于领域能力：`data/query.go`、`data/write.go`、`data/tx.go` 表达领域需要的数据能力，业务层不直接暴露 GORM、SQL、Redis 细节。
- HTTP 只是 adapter：request/response、status code、Gin context 留在 `http/` 包内，业务模型保持协议无关。
- Interface follows variation：只有真实变化点出现时才引入接口，并由使用方定义最小能力。
- AI 能力属于业务：`novaqwen` 只负责千问配置、HTTP Client 生命周期和 Chat Completions 调用；Prompt、结果解释、校验与降级属于调用方业务域，不预先创建泛化 AI 抽象。验证必须包含真实模型与代表性真实业务输入的运行证据。
- Reality-Driven Development：让真实运行结果驱动开发。优先运行系统、走真实用户路径并观察日志、指标和真实数据；没有新增运行证据的测试或二次审查不能单独证明可靠性。长期只保留极少量高价值冒烟测试。

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

## 现实驱动验证

验证首先来自真实运行：

1. 启动受影响的真实服务及其依赖。
2. 执行受影响的真实用户路径。
3. 检查响应、日志、指标和持久化数据。
4. 发现问题后修正实现，再次运行同一路径确认结果。
5. 回报运行方式、用户路径、观察结果和关键证据。

以下命令是辅助检查，不是完成验证的充分条件：

```bash
go test ./...
go vet ./...
```

## 快速索引

- `starter/config/novaconfig`
- `starter/http/novagin`
- `starter/gorm/novagorm`
- `starter/cache/novaredis`
- `starter/aliyun/novaoss`
- `starter/ai/novaqwen`
- `starter/realtime/novawebsocket`
- `examples/simple-app`
- `examples/best-practice-service`
- `docs/nova_engineering_best_practices.md`
- `docs/starter_conventions.md`
- `docs/starter_composition_matrix.md`
