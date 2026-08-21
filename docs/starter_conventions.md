# Starter 约定规范

本规范用于统一 `starter/*` 包的接口风格与使用方式。

## 设计目标

- 提供一致调用体验
- 减少跨 starter 心智切换
- 提高可读性与可维护性

## 统一约定

1. 初始化入口
- Starter 默认通过配置与首次访问完成初始化
- 实际资源创建尽量延迟到首次使用（lazy）

2. 获取实例
- 统一通过显式函数获取，如 `DB()/Client()/Router()`
- 返回值优先包含 `error`

3. 配置读取
- 统一通过 `novaconfig` 读取配置
- 配置 key 保持稳定、可预期
- 默认配置文件为 `config.yaml`
- 并非所有 starter 都必须读取配置；不读取配置的 starter 需要在专用说明中明确配置来源

4. 错误处理
- 错误信息应包含上下文与下一步动作提示
- 避免吞错，禁止静默失败

5. Reload 语义
- 提供 `Reload()` 的 starter 应保证可重复调用
- `Reload()` 仅负责重置/重载，不做业务副作用

## 新 starter 接入检查项

- 是否遵循配置驱动与 lazy 获取模式
- 是否有稳定配置 key 约定
- 是否有最小示例与基础测试
- 是否提供 `docs/starters/<starter>.md` 专用说明

## 配置 key 总览

| Starter | 配置来源 | 顶层 key | 说明 |
| --- | --- | --- | --- |
| `starter/config/novaconfig` | `config.yaml` | 无固定业务 key | 负责读取 YAML 配置，并支持点号读取嵌套 key |
| `starter/http/novagin` | `novaconfig` / 环境变量 | `http.port` | 端口优先级为 `OP_APP_PORT`、`http.port`、`PORT`、`8080` |
| `starter/cache/novaredis` | `novaconfig` | `redis` | 支持单实例与 `redis.instances.<name>` |
| `starter/aliyun/novaoss` | `novaconfig` | `aliyun.oss` | 支持单 Bucket 与 `aliyun.oss.instances.<name>` |
| `starter/realtime/novawebsocket` | 代码默认值 | 无 | 默认使用宽松 Upgrader，生产环境建议应用层收紧 `CheckOrigin` |

GORM 位于 [`orm/novagorm`](./orm/novagorm.md)，是 ORM 装配工具，不属于 `starter/*`。GORM 支持多实例；数据库类型和对应配置放在 `gorm.<name>` 下。只有一个实例时可以使用 `novagorm.DB()`；有多个实例时必须使用 `novagorm.Named("<name>").DB()`。

## 组合配置示例

```yaml
# starter/http/novagin
http:
  port: 8080

# orm/novagorm
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
  default: main
  instances:
    main:
      addr: localhost:6379
      db: 0

# starter/aliyun/novaoss
aliyun:
  oss:
    default: public
    instances:
      public:
        endpoint: https://oss-<region>.aliyuncs.com
        bucket: public-assets
        access_key_id: <runtime-secret>
        access_key_secret: <runtime-secret>
```

## 专用说明

- [`novaconfig`](./starters/novaconfig.md)
- [`novagin`](./starters/novagin.md)
- [`novaredis`](./starters/novaredis.md)
- [`novaoss`](./starters/novaoss.md)
- [`novawebsocket`](./starters/novawebsocket.md)
