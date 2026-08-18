# Starter 组合矩阵

## 场景 1：最小 API

- 必选：`glowconfig`、`glowhttp`
- 可选：无
- 适用：纯 HTTP 接口服务

## 场景 2：API + DB

- 必选：`glowconfig`、`glowhttp`、`glowmysql`
- 可选：`glowredis`
- 适用：读写型业务服务

## 场景 3：API + DB + Cache

- 必选：`glowconfig`、`glowhttp`、`glowmysql`、`glowredis`
- 可选：`glowwebsocket`
- 适用：高并发读写、缓存加速场景

## 场景 4：实时通信

- 必选：`glowconfig`、`glowhttp`、`glowwebsocket`
- 可选：`glowredis`、`glowmysql`
- 适用：推送、在线状态、实时协作

## 统一验证

```bash
go test ./...
go vet ./...
```
