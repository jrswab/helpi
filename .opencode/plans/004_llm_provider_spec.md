# LLM Provider Interface Specification

## 1. Overview

This document specifies the interface and behavior for the LLM provider layer in the Helpi Telegram bot. The provider layer enables routing user messages to multiple LLM backends (OpenAI, Anthropic, Ollama, OpenRouter, OpenCode Zen) through a unified interface.

## 2. Scope

### 2.1 In Scope

- Unified provider interface for all supported LLM backends
- Provider implementations for: OpenAI, Anthropic, Ollama, OpenRouter, OpenCode Zen
- Provider selection and routing based on user configuration
- Message formatting for each provider's API requirements
- Response extraction from each provider's API response format

### 2.2 Out of Scope

- Streaming responses to Telegram (handled by bot layer)
- Rate limiting or quota management
- Provider fallback or retry logic (handled by bot layer)
- Conversation memory (handled by session manager)

## 3. Design Goals

1. **Provider Agnostic**: Bot code should not need to know which provider is being used
2. **Configuration Driven**: Providers are enabled/disabled via config.yaml
3. **Standardized Messages**: All providers accept a common message format
4. **Error Transparency**: Errors include provider name and relevant details

## 4. Core Interfaces

### 4.1 Message Type

```go
type Message struct {
    Role    string // "system", "user", "assistant"
    Content string
}
```

The `Role` field MUST be one of:
- `"system"` - System prompts that define behavior
- `"user"` - User messages
- `"assistant"` - Previous model responses (for conversation context)

### 4.2 Provider Interface

```go
type Provider interface {
    // Name returns the provider identifier used in logs and errors
    Name() string
    
    // SendMessage sends messages to the LLM and returns the response content
    SendMessage(ctx context.Context, messages []Message) (string, error)
    
    // IsEnabled returns whether the provider is configured and available
    IsEnabled() bool
}
```

#### 4.2.1 Name() Specification

- MUST return a lowercase, kebab-case string
- MUST be one of: `"openai"`, `"anthropic"`, `"ollama"`, `"openrouter"`, `"opencode"`
- MUST NOT return different values for the same provider across calls

#### 4.2.2 SendMessage() Specification

**Input:**
- `ctx`: Context with timeout (timeout value is specified in configuration)
- `messages`: Slice of Message structs in chronological order (oldest first)

**Output:**
- Returns the assistant's response content as a string
- Returns error if:
  - API request fails (network error, timeout)
  - API returns an error response
  - Response cannot be parsed

**Error Format:**
- All errors MUST include the provider name via `fmt.Errorf` with `%s: %w, provider.Name(), err`
- This allows callers to identify which provider failed

#### 4.2.3 IsEnabled() Specification

- Returns `true` if the provider is:
  - Enabled in configuration (`enabled: true` in config.yaml)
  - Has valid credentials (API key present in .env)
- Returns `false` otherwise

### 4.3 Router Interface

```go
type Router interface {
    // GetProvider returns the active provider based on configuration
    // Returns error if no provider is enabled or available
    GetProvider() (Provider, error)
    
    // SendMessage routes a message through the active provider
    SendMessage(ctx context.Context, messages []Message) (string, error)
}
```

#### 4.3.1 GetProvider() Specification

- Returns the provider marked as default in configuration
- If no default is set, returns the first enabled provider
- Returns error with descriptive message if no provider is enabled

#### 4.3.2 SendMessage() Specification

- Convenience method that calls GetProvider() and then sends to that provider
- Propagates errors from GetProvider() or from the provider's SendMessage()

## 5. Supported Providers

### 5.1 OpenAI

| Attribute | Value |
|-----------|-------|
| Interface Name | OpenAI Provider |
| Config Key | `providers.openai` |
| Environment Variable | `OPENAI_API_KEY` |
| API Base URL | `https://api.openai.com/v1` |
| Default Model | Configurable via `default_model` (e.g., `"gpt-4o"`) |
| SDK | `github.com/openai/openai-go/v3` |

**Message Mapping:**

| Input Role | API Field |
|------------|-----------|
| `"system"` | `openai.SystemMessage()` |
| `"user"` | `openai.UserMessage()` |
| `"assistant"` | `openai.AssistantMessage()` |

**Response Extraction:**
- Access `response.Choices[0].Message.Content`
- If choices is empty, return empty string (no error)

**Edge Cases:**
- If API key is missing: `IsEnabled()` returns `false`
- If model is not specified: use default model from config
- If API returns rate limit error: propagate with provider name

### 5.2 Anthropic

| Attribute | Value |
|-----------|-------|
| Interface Name | Anthropic Provider |
| Config Key | `providers.anthropic` |
| Environment Variable | `ANTHROPIC_API_KEY` |
| API Base URL | `https://api.anthropic.com` |
| Default Model | Configurable via `default_model` (e.g., `"claude-sonnet-4-5-20250514"`) |
| SDK | `github.com/anthropics/anthropic-sdk-go` |

**Message Mapping:**

| Input Role | API Field |
|------------|-----------|
| `"system"` | `anthropic.NewSystemMessage()` with `anthropic.NewTextBlock()` |
| `"user"` | `anthropic.NewUserMessage()` with `anthropic.NewTextBlock()` |
| `"assistant"` | Converted to user message with `role: assistant` |

**Request Parameters:**
- `MaxTokens`: MUST be set to a positive integer (default: 1024)
- `Model`: MUST match config value exactly

**Response Extraction:**
- Access `message.Content[0].GetText()`
- Content is a slice; concatenate all text blocks if multiple present

**Edge Cases:**
- If API key is missing: `IsEnabled()` returns `false`
- If `MaxTokens` exhausted: response may contain `stop_reason: "max_tokens"`, still return content
- System messages: Anthropic has a dedicated system parameter, not in messages array

### 5.3 Ollama

| Attribute | Value |
|-----------|-------|
| Interface Name | Ollama Provider |
| Config Key | `providers.ollama` |
| Environment Variable | None (uses local server) |
| API Base URL | From config: `OLLAMA_BASE_URL` (default: `http://localhost:11434/v1`) |
| Default Model | Configurable via `default_model` (e.g., `"llama3"`) |
| SDK | `github.com/openai/openai-go/v3` (with custom base URL) |

**Message Mapping:** Same as OpenAI

**Configuration:**
- API Key: Use placeholder value (e.g., `"ollama"`) - required by SDK but not validated
- Base URL: Read from config, default to `http://localhost:11434/v1`

**Response Extraction:** Same as OpenAI

**Edge Cases:**
- If Ollama server is not running: connection timeout error
- If model is not available locally: API returns error
- If base URL is unreachable: connection error

### 5.4 OpenRouter

| Attribute | Value |
|-----------|-------|
| Interface Name | OpenRouter Provider |
| Config Key | `providers.openrouter` |
| Environment Variable | `OPENROUTER_API_KEY` |
| API Base URL | `https://openrouter.ai/api/v1` |
| Default Model | Configurable via `default_model` (e.g., `"openai/gpt-4o"`) |
| SDK | `github.com/openai/openai-go/v3` (with custom base URL) |

**Message Mapping:** Same as OpenAI

**Required Headers:**
- `HTTP-Referer`: Optional, for ranking (can be empty or app URL)
- `X-Title`: Optional, set to `"Helpi"`

**Model Format:**
- Must use format `"provider/model-name"` (e.g., `"openai/gpt-4o"`, `"anthropic/claude-sonnet-4-20250514"`)
- The default_model in config MUST use this format

**Response Extraction:** Same as OpenAI

**Edge Cases:**
- If API key is missing: `IsEnabled()` returns `false`
- Invalid model format: API returns error
- Credit exhaustion: API returns error with details

### 5.5 OpenCode Zen

| Attribute | Value |
|-----------|-------|
| Interface Name | OpenCode Provider |
| Config Key | `providers.opencode` |
| Environment Variable | `OPENCODE_API_KEY` |
| API Base URL | `https://opencode.ai/zen/v1` |
| Default Model | Configurable via `default_model` (e.g., `"opencode/big-pickle"`) |
| SDK | `github.com/openai/openai-go/v3` (with custom base URL) |

**Message Mapping:** Same as OpenAI

**Model Format:**
- Must use format `"opencode/model-name"` (e.g., `"opencode/big-pickle"`, `"opencode/claude-3-5-haiku"`)

**Response Extraction:** Same as OpenAI

**Edge Cases:**
- If API key is missing: `IsEnabled()` returns `false`
- Invalid model name: API returns error

## 6. Configuration Integration

### 6.1 Required Config Structure

The LLM layer reads from the existing `config.yaml` structure:

```yaml
providers:
  openai:
    enabled: true
    default_model: "gpt-4o"
  anthropic:
    enabled: false
  openrouter:
    enabled: true
    default_model: "openai/gpt-4o"
  opencode:
    enabled: true
    default_model: "opencode/big-pickle"
  ollama:
    enabled: false

memory:
  path: "./data/sessions"
  max_messages: 50
```

### 6.2 Environment Variables

| Config Key | Environment Variable | Required When |
|------------|---------------------|---------------|
| `providers.openai` | `OPENAI_API_KEY` | enabled = true |
| `providers.anthropic` | `ANTHROPIC_API_KEY` | enabled = true |
| `providers.openrouter` | `OPENROUTER_API_KEY` | enabled = true |
| `providers.opencode` | `OPENCODE_API_KEY` | enabled = true |
| `providers.ollama` | (none) | enabled = true |
| `providers.ollama` | `OLLAMA_BASE_URL` | enabled = true (optional, has default) |

### 6.3 Default Provider Selection

- The provider with `enabled: true` AND explicitly marked as default in config
- If no default is marked, the first enabled provider in config order
- If no providers are enabled, router returns error: "no LLM provider enabled"

## 7. Error Handling

### 7.1 Error Types

All errors MUST be wrapped with provider context:

```go
return "", fmt.Errorf("%s: %w", p.Name(), err)
```

### 7.2 Error Scenarios

| Scenario | Behavior |
|----------|----------|
| No provider enabled | Return error: "no LLM provider enabled" |
| API key missing | Provider returns `IsEnabled() = false` |
| Network timeout | Return wrapped error with timeout message |
| Invalid model | Return API error message |
| Rate limited | Return wrapped error with rate limit message |
| Response parse failure | Return error: "failed to parse response from {provider}" |

### 7.3 No Silent Failures

- Empty response with no error: Return empty string (valid for some prompts)
- Non-empty error: MUST return error, never empty string

## 8. Factory Pattern

### 8.1 Provider Factory

```go
// NewProvider creates a provider instance based on configuration
func NewProvider(config *config.Config, providerType string) (Provider, error)
```

**Parameters:**
- `config`: Full application configuration
- `providerType`: One of `"openai"`, `"anthropic"`, `"ollama"`, `"openrouter"`, `"opencode"`

**Returns:**
- Initialized provider or error if configuration is invalid

### 8.2 Router Factory

```go
// NewRouter creates a router with all configured providers
func NewRouter(config *config.Config) (Router, error)
```

**Parameters:**
- `config`: Full application configuration

**Returns:**
- Router with all enabled providers registered, or error if no providers enabled

## 9. File Structure

```
internal/llm/
├── router.go       # Router interface and implementation
├── factory.go      # Provider and router factory functions
├── openai.go       # OpenAI provider implementation
├── anthropic.go    # Anthropic provider implementation
├── ollama.go       # Ollama provider implementation
├── openrouter.go   # OpenRouter provider implementation
└── opencode.go     # OpenCode Zen provider implementation
```

## 10. Testing Requirements

### 10.1 Unit Tests

- Each provider MUST have unit tests for message formatting
- Response extraction MUST be tested with mock responses
- Error handling MUST be tested for each error scenario

### 10.2 Integration Tests (Optional)

- Each provider SHOULD have an integration test if credentials are available
- Use environment variables to skip tests when credentials are missing

## 11. Security Considerations

- API keys MUST only be read from environment variables, never hardcoded
- Error messages MUST NOT expose API keys or full request details
- Provider selection MUST respect configuration, never bypass enabled checks