# novaqwen

`starter/ai/novaqwen` 是阿里云千问基础设施 Starter。它从 Nova 配置读取凭据、地址、模型和超时，延迟创建并统一管理 HTTP Client；业务代码只组装消息并解释结果。

## 配置

```yaml
ai:
  qwen:
    api_key: <runtime-secret>
    endpoint: https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions
    model: qwen-plus
    timeout_seconds: 15
```

`endpoint`、`model` 和 `timeout_seconds` 均有默认值，`api_key` 必须由运行时配置或环境变量提供。配置值优先于环境变量。

环境变量兼容：

- 凭据：`QWEN_API_KEY`、`DASHSCOPE_API_KEY`
- 地址：`QWEN_API_BASE`、`QWEN_ENDPOINT`、`DASHSCOPE_API_BASE`、`DASHSCOPE_BASE_URL`、`DASHSCOPE_ENDPOINT`
- 模型：`QWEN_MODEL`、`DASHSCOPE_MODEL`
- 超时秒数：`QWEN_TIMEOUT_SECONDS`、`DASHSCOPE_TIMEOUT_SECONDS`

## 使用

```go
package analysis

import (
	"context"

	"github.com/luaxlou/nova/starter/ai/novaqwen"
)

func generate(ctx context.Context, prompt string) (string, error) {
	response, err := novaqwen.Chat(ctx, novaqwen.Request{Messages: []novaqwen.Message{
		{Role: "system", Content: "按要求生成结果。"},
		{Role: "user", Content: prompt},
	}})
	if err != nil {
		return "", err
	}
	return response.FirstContent()
}
```

应用关闭时调用 `novaqwen.CloseAll()`。切换配置后先调用 `novaconfig.Reload()`，再调用 `novaqwen.Reload()`。

## 边界

- Starter 只负责阿里云千问的配置、HTTP Client 生命周期和 Chat Completions 调用。
- Prompt、业务输入输出、结果解释与降级策略属于调用方业务领域。
- Starter 不提供通用 AI、Agent、风险分析或其他业务抽象。
