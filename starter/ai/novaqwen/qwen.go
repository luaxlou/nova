// Package novaqwen provides a config-driven Alibaba Cloud Qwen client.
package novaqwen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/luaxlou/nova/internal/registry"
	"github.com/luaxlou/nova/starter/config/novaconfig"
)

const (
	defaultEndpoint = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	defaultModel    = "qwen-plus"
	defaultTimeout  = 15 * time.Second
	maxResponseSize = 8 << 20
	maxAttempts     = 3
	singletonName   = "single"
)

var (
	// ErrInvalidRequest identifies a locally rejected Qwen request.
	ErrInvalidRequest = errors.New("novaqwen: invalid request")
	// ErrInvalidResponse identifies a malformed or incomplete Qwen response.
	ErrInvalidResponse = errors.New("novaqwen: invalid response")

	initialized bool
	initMu      sync.Mutex
	reg         = registry.New[*Client]()
)

// Client owns the HTTP transport and runtime configuration for Qwen.
// Applications obtain it through Open rather than constructing it directly.
type Client struct {
	apiKey         string
	endpoint       string
	model          string
	enableThinking bool
	httpClient     *http.Client
}

// Message is one text message in a Qwen conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Request is a non-streaming Qwen chat completion request. The model comes
// from Starter configuration, not from business code.
type Request struct {
	Messages []Message
}

// Response is a non-streaming Qwen chat completion response.
type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice is one Qwen completion candidate.
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage reports token consumption returned by Qwen.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// APIError is returned for a non-2xx Qwen response.
type APIError struct {
	StatusCode int
	Status     string
	Code       string
	Message    string
	RequestID  string
}

// RetryError reports that a transient Qwen request failure exhausted all
// Starter-owned attempts. The final infrastructure error remains available
// through errors.Is and errors.As.
type RetryError struct {
	Attempts int
	Err      error
}

func (e *RetryError) Error() string {
	return fmt.Sprintf("novaqwen: request failed after %d/%d attempts: %v", e.Attempts, maxAttempts, e.Err)
}

func (e *RetryError) Unwrap() error { return e.Err }

func (e *APIError) Error() string {
	parts := []string{fmt.Sprintf("novaqwen: API request failed: status=%s", e.Status)}
	if e.Code != "" {
		parts = append(parts, "code="+e.Code)
	}
	if e.Message != "" {
		parts = append(parts, "message="+e.Message)
	}
	if e.RequestID != "" {
		parts = append(parts, "request_id="+e.RequestID)
	}
	return strings.Join(parts, " ")
}

// Open returns the lazily initialized Qwen singleton.
func Open() (*Client, error) {
	if err := ensureInit(); err != nil {
		return nil, err
	}
	return reg.Get().Get()
}

// Chat sends a request through the Qwen singleton.
func Chat(ctx context.Context, request Request) (Response, error) {
	client, err := Open()
	if err != nil {
		return Response{}, err
	}
	return client.Chat(ctx, request)
}

// Chat performs a non-streaming OpenAI-compatible Qwen chat completion.
func (c *Client) Chat(ctx context.Context, request Request) (Response, error) {
	if err := validateRequest(ctx, request); err != nil {
		return Response{}, err
	}

	body, err := json.Marshal(struct {
		Model          string    `json:"model"`
		Messages       []Message `json:"messages"`
		Stream         bool      `json:"stream"`
		EnableThinking bool      `json:"enable_thinking"`
	}{Model: c.model, Messages: request.Messages, Stream: false, EnableThinking: c.enableThinking})
	if err != nil {
		return Response{}, fmt.Errorf("novaqwen: encode request: %w", err)
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		response, err := c.send(ctx, body)
		if err == nil {
			return response, nil
		}
		if ctx.Err() != nil {
			return Response{}, fmt.Errorf("novaqwen: request stopped at attempt %d/%d: %w", attempt, maxAttempts, ctx.Err())
		}
		if !transientRequestError(err) {
			return Response{}, err
		}
		if attempt == maxAttempts {
			return Response{}, &RetryError{Attempts: attempt, Err: err}
		}
		if err := waitForRetry(ctx, time.Duration(attempt)*time.Second); err != nil {
			return Response{}, fmt.Errorf("novaqwen: request stopped after attempt %d/%d: %w", attempt, maxAttempts, err)
		}
	}
	return Response{}, errors.New("novaqwen: request attempts exhausted")
}

func (c *Client) send(ctx context.Context, body []byte) (Response, error) {
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("novaqwen: create request: %w", err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return Response{}, fmt.Errorf("novaqwen: send request: %w", err)
	}
	defer httpResponse.Body.Close()

	raw, err := readResponse(httpResponse.Body)
	if err != nil {
		return Response{}, err
	}
	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		return Response{}, newAPIError(httpResponse, raw)
	}

	var response Response
	if err := json.Unmarshal(raw, &response); err != nil {
		return Response{}, fmt.Errorf("%w: decode JSON: %v", ErrInvalidResponse, err)
	}
	if len(response.Choices) == 0 {
		return Response{}, fmt.Errorf("%w: choices are empty", ErrInvalidResponse)
	}
	return response, nil
}

func transientRequestError(err error) bool {
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var networkError net.Error
	if errors.As(err, &networkError) {
		return true
	}
	var apiError *APIError
	return errors.As(err, &apiError) && (apiError.StatusCode == http.StatusTooManyRequests || apiError.StatusCode >= http.StatusInternalServerError)
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// FirstContent returns the first non-empty text completion.
func (r Response) FirstContent() (string, error) {
	for _, choice := range r.Choices {
		if content := strings.TrimSpace(choice.Message.Content); content != "" {
			return content, nil
		}
	}
	return "", fmt.Errorf("%w: completion content is empty", ErrInvalidResponse)
}

// Reload re-reads ai.qwen from novaconfig and replaces the cached client.
func Reload() error {
	initMu.Lock()
	defer initMu.Unlock()
	if err := reg.CloseAll(); err != nil {
		return fmt.Errorf("reload ai.qwen config: close current client: %w", err)
	}
	if err := configureFromCurrentConfig(); err != nil {
		return fmt.Errorf("reload ai.qwen config: %w", err)
	}
	return nil
}

// CloseAll releases the cached Qwen client and its idle HTTP connections.
func CloseAll() error {
	if err := ensureInit(); err != nil {
		return err
	}
	return reg.CloseAll()
}

// Close releases idle HTTP connections owned by this client.
func (c *Client) Close() error {
	if c != nil && c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
	}
	return nil
}

func validateRequest(ctx context.Context, request Request) error {
	if ctx == nil {
		return fmt.Errorf("%w: context is nil", ErrInvalidRequest)
	}
	if len(request.Messages) == 0 {
		return fmt.Errorf("%w: messages are required", ErrInvalidRequest)
	}
	for index, message := range request.Messages {
		if strings.TrimSpace(message.Role) == "" {
			return fmt.Errorf("%w: messages[%d].role is required", ErrInvalidRequest, index)
		}
		if strings.TrimSpace(message.Content) == "" {
			return fmt.Errorf("%w: messages[%d].content is required", ErrInvalidRequest, index)
		}
	}
	return nil
}

func readResponse(body io.Reader) ([]byte, error) {
	raw, err := io.ReadAll(io.LimitReader(body, maxResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("novaqwen: read response: %w", err)
	}
	if len(raw) > maxResponseSize {
		return nil, fmt.Errorf("%w: response exceeds %d bytes", ErrInvalidResponse, maxResponseSize)
	}
	return raw, nil
}

func newAPIError(response *http.Response, raw []byte) *APIError {
	result := &APIError{
		StatusCode: response.StatusCode,
		Status:     response.Status,
		RequestID:  strings.TrimSpace(response.Header.Get("x-request-id")),
	}
	var payload struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id"`
		Error     struct {
			Code      string `json:"code"`
			Message   string `json:"message"`
			RequestID string `json:"request_id"`
		} `json:"error"`
	}
	if json.Unmarshal(raw, &payload) == nil {
		result.Code = firstNonEmpty(payload.Error.Code, payload.Code)
		result.Message = firstNonEmpty(payload.Error.Message, payload.Message)
		result.RequestID = firstNonEmpty(result.RequestID, payload.Error.RequestID, payload.RequestID)
	}
	if result.Message == "" {
		result.Message = truncate(strings.Join(strings.Fields(string(raw)), " "), 1024)
	}
	return result
}

func ensureInit() error {
	if initialized {
		return nil
	}
	initMu.Lock()
	defer initMu.Unlock()
	if initialized {
		return nil
	}
	return configureFromCurrentConfig()
}

func configureFromCurrentConfig() error {
	config, err := loadConfig()
	if err != nil {
		reg.Configure("", map[string]registry.Builder[*Client]{})
		initialized = false
		return err
	}
	reg.Configure(singletonName, map[string]registry.Builder[*Client]{
		singletonName: func(_ string) (*Client, error) {
			return newClient(config)
		},
	})
	initialized = true
	return nil
}

type qwenConfig struct {
	Endpoint       string
	APIKey         string
	Model          string
	EnableThinking bool
	Timeout        time.Duration
}

func loadConfig() (qwenConfig, error) {
	config := qwenConfig{
		Endpoint:       firstNonEmpty(novaconfig.GetString("ai.qwen.endpoint"), defaultEndpoint),
		APIKey:         strings.TrimSpace(novaconfig.GetString("ai.qwen.api_key")),
		Model:          firstNonEmpty(novaconfig.GetString("ai.qwen.model"), defaultModel),
		EnableThinking: novaconfig.GetBool("ai.qwen.enable_thinking"),
		Timeout:        defaultTimeout,
	}

	timeout := novaconfig.GetInt("ai.qwen.timeout_seconds")
	if timeout > 0 {
		config.Timeout = time.Duration(timeout) * time.Second
	}

	if config.APIKey == "" {
		return qwenConfig{}, fmt.Errorf("ai.qwen api_key is required")
	}
	endpoint, err := url.ParseRequestURI(config.Endpoint)
	if err != nil || endpoint.Scheme == "" || endpoint.Host == "" || (endpoint.Scheme != "http" && endpoint.Scheme != "https") {
		return qwenConfig{}, fmt.Errorf("ai.qwen endpoint is invalid")
	}
	return config, nil
}

func newClient(config qwenConfig) (*Client, error) {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("novaqwen: default HTTP transport is unsupported")
	}
	return &Client{
		apiKey:         config.APIKey,
		endpoint:       config.Endpoint,
		model:          config.Model,
		enableThinking: config.EnableThinking,
		httpClient: &http.Client{
			Transport: transport.Clone(),
			Timeout:   config.Timeout,
		},
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
