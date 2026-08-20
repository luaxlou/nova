# Starter 组合矩阵

## 场景 1：最小 API

- 必选：`novaconfig`、`novagin`
- 可选：无
- 适用：纯 HTTP 接口服务

## 场景 2：API + DB

- 必选：`novaconfig`、`novagin`、`novamysql`
- 可选：`novagorm`、`novaredis`
- 适用：读写型业务服务

## 场景 3：API + DB + Cache

- 必选：`novaconfig`、`novagin`、`novamysql`、`novaredis`
- 可选：`novagorm`、`novawebsocket`
- 适用：高并发读写、缓存加速场景

## 场景 4：实时通信

- 必选：`novaconfig`、`novagin`、`novawebsocket`
- 可选：`novaredis`、`novamysql`
- 适用：推送、在线状态、实时协作

## 场景 5：对象存储

- 必选：`novaconfig`、`starter/aliyun/novaoss`
- 可选：`novagin`、`novamysql`
- 适用：使用 Alibaba Cloud OSS 保存或读取对象；访问密钥通过运行时配置或密钥管理系统提供，不写入源代码

## 统一验证

```bash
go test ./...
go vet ./...
```
