# Nova Examples

本目录提供面向应用侧 starter 的可运行示例。

## 示例清单

- `simple-app`：最小配置读取示例
- `minimal-api`：最小 HTTP 服务示例（`novaconfig` + `novahttp`）
- `api-db-cache`：HTTP + MySQL + Redis 组合示例
- `api-websocket`：HTTP + WebSocket 组合示例

## 运行提示

示例默认读取项目根目录下的 `config.yaml`。
请按示例中使用的 starter 准备对应配置字段。

统一验证命令：

```bash
go test ./...
go vet ./...
```
