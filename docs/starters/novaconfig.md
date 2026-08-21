# novaconfig

`starter/config/novaconfig` 是 Nova 的统一配置入口。默认读取当前工作目录下的 `config.yaml`，并支持使用点号读取嵌套 key。

## 配置文件

```yaml
app:
  name: demo
  max_connections: 100
```

## 最小用法

```go
package main

import (
	"fmt"

	"github.com/luaxlou/nova/starter/config/novaconfig"
)

func main() {
	fmt.Println(novaconfig.GetString("app.name"))
	fmt.Println(novaconfig.GetInt("app.max_connections"))
}
```

## 约定

- 默认配置文件名为 `config.yaml`
- 需要自定义路径时，在首次读取前调用 `novaconfig.SetConfigPath(path)`
- 配置变更后可调用 `novaconfig.Reload()` 重新加载
- 其他需要配置的 starter 统一通过 `novaconfig` 读取
