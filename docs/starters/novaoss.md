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
    default: public
    instances:
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

## 约定

- 配置挂在 `aliyun.oss` 下
- 默认 Bucket 使用 `novaoss.Bucket()` 或 `novaoss.Get().Bucket()`
- 命名 Bucket 使用 `novaoss.Named("private").Bucket()`
- 访问密钥和临时令牌必须来自运行时配置或密钥管理系统，禁止写入源代码或提交到仓库
