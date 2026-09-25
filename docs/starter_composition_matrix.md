# Starter 组合矩阵

## 场景 1：最小 API

- 必选：`novaconfig`、`novagin`
- 可选：无
- 适用：纯 HTTP 接口服务

## 场景 2：API + DB

- 必选：`novaconfig`、`novagin`、`starter/gorm/novagorm`
- 可选：`novaredis`
- 适用：使用 MySQL 或 PostgreSQL 的读写型业务服务；PostgreSQL 18 是当前真实验证基线

## 场景 3：API + DB + Cache

- 必选：`novaconfig`、`novagin`、`starter/gorm/novagorm`、`novaredis`
- 可选：`novawebsocket`
- 适用：高并发读写、缓存加速场景

## 场景 4：实时通信

- 必选：`novaconfig`、`novagin`、`novawebsocket`
- 可选：`novaredis`、`starter/gorm/novagorm`
- 适用：推送、在线状态、实时协作

## 场景 5：对象存储

- 必选：`novaconfig`、`starter/aliyun/novaoss`
- 可选：`novagin`、`starter/gorm/novagorm`
- 适用：使用 Alibaba Cloud OSS 保存或读取对象；访问密钥通过运行时配置或密钥管理系统提供，不写入源代码

## 统一验证

常规变更运行：

```bash
go test ./...
go vet ./...
```

PostgreSQL driver 发布前还必须连接专用 PostgreSQL 18 测试库运行：

```bash
POSTGRES_TEST_DSN='<runtime-secret>' go test -tags=postgres_integration ./starter/gorm/novagorm -run TestPostgres18Integration -count=1 -v
```
