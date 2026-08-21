# novaoss

`starter/aliyun/novaoss` 基于 Alibaba Cloud OSS 官方 Go SDK，提供 lazy、named 的 Bucket 访问。

## 单实例配置

```yaml
aliyun:
  oss:
    endpoint: https://oss-<region>.aliyuncs.com
    bucket: <bucket>
    access_key_id: <runtime-secret>
    access_key_secret: <runtime-secret>
    security_token: <optional-runtime-secret>
```

## 多实例配置

```yaml
aliyun:
  oss:
    public:
      endpoint: https://oss-<region>.aliyuncs.com
      bucket: public-assets
      access_key_id: <runtime-secret>
      access_key_secret: <runtime-secret>
    private:
      endpoint: https://oss-<region>.aliyuncs.com
      bucket: private-assets
      access_key_id: <runtime-secret>
      access_key_secret: <runtime-secret>
```

## 最小用法

```go
package main

import "github.com/luaxlou/nova/starter/aliyun/novaoss"

func main() {
	bucket, err := novaoss.Bucket()
	if err != nil {
		panic(err)
	}

	_ = bucket
}
```

## 重载配置

应用切换配置路径或更新配置后，先调用 `novaconfig.Reload()`，再调用 `novaoss.Reload()`。后者会重新读取当前 `aliyun.oss` 配置、替换全部命名定义并清除已缓存的 Bucket；Bucket 仍在下一次访问时才创建。

```go
novaconfig.SetConfigPath("/etc/my-app/config.yaml")
if err := novaconfig.Reload(); err != nil {
	return err
}
if err := novaoss.Reload(); err != nil {
	return err
}
```

`novaoss.Reload()` 会返回带配置上下文的错误。重载失败时，旧的缓存定义会被清除，后续访问不会静默继续使用旧的 Bucket 配置。

## 约定

- 配置挂在 `aliyun.oss` 下
- 只有一个实例时可以使用 `novaoss.Bucket()`
- 有多个实例时必须使用 `novaoss.Named("<name>").Bucket()`
- 访问密钥和临时令牌必须来自运行时配置或密钥管理系统，禁止写入源代码或提交到仓库
