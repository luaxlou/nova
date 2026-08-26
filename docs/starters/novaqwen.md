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

`endpoint`、`model` 和 `timeout_seconds` 均有代码默认值，`api_key` 必须通过 `novaconfig` 的 `ai.qwen.api_key` 提供。Starter 不直接读取环境变量。

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

## 瞬时故障处理

`novaqwen` 对连接错误、`EOF`、响应体意外中断、HTTP 429 和 HTTP 5xx 最多尝试三次，重试前分别等待一秒和两秒。每次尝试都会重新创建 POST 请求；上下文取消会立即终止等待和后续调用，参数错误、鉴权错误及其他非瞬时错误不会重试。

三次尝试仍失败时返回 `*novaqwen.RetryError`。调用方可通过 `errors.As` 读取 `Attempts`，并继续通过 `errors.Is` 或 `errors.As` 检查最后一次原始错误。Starter 保持静默，不记录 Prompt、响应、凭据或业务证据。

## 边界

- Starter 只负责阿里云千问的配置、HTTP Client 生命周期和 Chat Completions 调用。
- Prompt、业务输入输出、结果解释与降级策略属于调用方业务领域。
- Starter 不提供通用 AI、Agent、风险分析或其他业务抽象。
