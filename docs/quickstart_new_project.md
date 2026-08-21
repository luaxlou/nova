# 新项目引入 Nova（5 步流程卡）

## 目标

在新 Go 服务中快速引入 nova，并形成统一初始化范式，让项目从第一天就享受 nova 设计哲学收益。

## 前置输入

- Go 1.24+
- 目标服务仓库
- `config.yaml` 基础配置

## 5 步流程

1. 接入依赖
- 执行：`go get github.com/luaxlou/nova`

2. 建立初始化骨架
- 在 `main.go` 中先接入 `novaconfig` 与 `novagin`
- 按需声明 `orm/novagorm`、`novaredis`、`novawebsocket`

3. 统一路由与依赖注入入口
- 所有启动逻辑集中在单一入口（建议 `cmd/<app>/main.go`）

4. 补充可运行配置
- 准备 `config.yaml`
- 明确 `gorm.driver`、`gorm.mysql.dsn`、`redis.addr`、`PORT` 等字段来源

5. 运行与验证
- 启动服务并完成最小请求验证
- 执行：`go test ./...`、`go vet ./...`

## 期望输出

- 可运行服务入口
- 清晰的 starter 组合说明
- 验证结果（测试与 vet）
