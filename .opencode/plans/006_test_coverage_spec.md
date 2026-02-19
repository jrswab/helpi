# Test Coverage Specification

## 1. Overview

This document specifies the test requirements for all untested packages in the Helpi project. The goal is to achieve comprehensive unit test coverage for all core functionality.

## 2. Current Coverage Status

| Package | Current Coverage | Status |
|---------|------------------|--------|
| `internal/config` | 72.2% | ✓ Has tests |
| `internal/llm` | 0.0% | ✗ No tests |
| `internal/session` | 0.0% | ✗ No tests |
| `internal/bot` | 0.0% | ✗ No tests |
| `cmd/bot` | 0.0% | ✗ No tests |
| `cmd/setup` | 0.0% | ✗ No tests |

## 3. Scope

### 3.1 In Scope

- Unit tests for all packages except `cmd/bot` (entry point, limited testability)
- Integration tests should NOT be included (require live API keys)

### 3.2 Out of Scope

- Integration tests requiring live LLM API keys
- End-to-end tests with actual Telegram bot
- Performance benchmarks

## 4. Test Files to Create

### 4.1 `internal/llm/router_test.go`

**Purpose**: Test Router interface implementation

**Functions to Test**:
| Function | Scenario | Expected |
|----------|----------|----------|
| `GetProvider()` | Default provider enabled | Returns default provider |
| `GetProvider()` | Default disabled, another enabled | Falls back to first enabled |
| `GetProvider()` | No providers enabled | Returns error |
| `SendMessage()` | Valid provider | Returns response |
| `SendMessage()` | No provider | Returns error |

### 4.2 `internal/llm/factory_test.go`

**Purpose**: Test factory functions

**Functions to Test**:
| Function | Scenario | Expected |
|----------|----------|----------|
| `NewProvider()` | "openai" | Returns OpenAI provider |
| `NewProvider()` | "anthropic" | Returns Anthropic provider |
| `NewProvider()` | "ollama" | Returns Ollama provider |
| `NewProvider()` | "unknown" | Returns error |
| `NewRouter()` | Multiple providers | Returns router with all |
| `NewRouter()` | No providers | Returns error |

### 4.3 `internal/llm/openai_test.go`

**Purpose**: Test OpenAI provider

**Functions to Test**:
| Method | Scenario | Expected |
|--------|----------|----------|
| `Name()` | Always | Returns "openai" |
| `IsEnabled()` | Enabled + API key | Returns true |
| `IsEnabled()` | Disabled | Returns false |
| `IsEnabled()` | No API key | Returns false |
| `SendMessage()` | Disabled | Returns error |
| `SendMessage()` | Network error | Returns wrapped error |

### 4.4 `internal/llm/anthropic_test.go`

**Purpose**: Test Anthropic provider

| Method | Scenario | Expected |
|--------|----------|----------|
| `Name()` | Always | Returns "anthropic" |
| `IsEnabled()` | Enabled + API key | Returns true |
| `IsEnabled()` | Disabled | Returns false |
| `SendMessage()` | Disabled | Returns error |

### 4.5 `internal/llm/ollama_test.go`

**Purpose**: Test Ollama provider

| Method | Scenario | Expected |
|--------|----------|----------|
| `Name()` | Always | Returns "ollama" |
| `IsEnabled()` | Enabled | Returns true |
| `IsEnabled()` | Disabled | Returns false |

### 4.6 `internal/llm/openrouter_test.go`

**Purpose**: Test OpenRouter provider

| Method | Scenario | Expected |
|--------|----------|----------|
| `Name()` | Always | Returns "openrouter" |
| `IsEnabled()` | Enabled + API key | Returns true |

### 4.7 `internal/llm/opencode_test.go`

**Purpose**: Test OpenCode provider

| Method | Scenario | Expected |
|--------|----------|----------|
| `Name()` | Always | Returns "opencode" |
| `IsEnabled()` | Enabled + API key | Returns true |

### 4.8 `internal/session/manager_test.go`

**Purpose**: Test session manager

| Method | Scenario | Expected |
|--------|----------|----------|
| `NewManager()` | Empty path | Uses default "./data/sessions" |
| `NewManager()` | MaxMessages = 0 | Uses default 50 |
| `NewManager()` | Invalid directory | Returns error |
| `Get()` | No session file | Returns empty slice, nil |
| `Get()` | Valid JSON | Returns parsed messages |
| `Get()` | Corrupted JSON | Returns error |
| `Save()` | New session | Creates JSON file |
| `Save()` | Exceeds maxMessages | Truncates to max |
| `Save()` | Marshal error | Returns error |
| `Save()` | Write error | Returns error |
| `Delete()` | Existing file | Removes file, nil |
| `Delete()` | Non-existent | Returns nil (graceful) |

### 4.9 `internal/bot/auth_test.go`

**Purpose**: Test auth middleware

| Function | Scenario | Expected |
|----------|----------|----------|
| `AuthMiddleware()` | Authorized user | Passes to next |
| `AuthMiddleware()` | Unauthorized user | Sends access denied |
| `AuthMiddleware()` | Empty allowedUsers | Returns true (dev mode) |
| `extractUserID()` | Message | Returns Message.From.ID |
| `extractUserID()` | CallbackQuery | Returns CallbackQuery.From.ID |
| `extractUserID()` | EditedMessage | Returns EditedMessage.From.ID |
| `extractUserID()` | Empty update | Returns 0 |
| `getChatID()` | Message | Returns Message.Chat.ID |
| `getChatID()` | CallbackQuery | Returns appropriate chat ID |

### 4.10 `internal/bot/handlers_test.go`

**Purpose**: Test command and message handlers

| Handler | Scenario | Expected |
|---------|----------|----------|
| `StartHandler` | Authorized | Sends welcome message |
| `HelpHandler` | Authorized | Sends help text |
| `MyIDHandler` | Authorized | Sends user ID |
| `ModelHandler` | Provider available | Sends provider info |
| `ModelHandler` | No provider | Sends error |
| `ClearHandler` | Delete success | Sends confirmation |
| `ClearHandler` | Delete fails | Sends error message |
| `TextMessageHandler` | Session load error | Sends error |
| `TextMessageHandler` | LLM success | Sends response |
| `TextMessageHandler` | LLM error | Sends error message |

### 4.11 `cmd/setup/main_test.go`

**Purpose**: Test setup wizard functions

| Function | Scenario | Expected |
|----------|----------|----------|
| `promptToken()` | Empty input | Returns current value |
| `promptToken()` | Non-empty input | Returns input |
| `promptModel()` | Empty input | Returns default |
| `promptModel()` | Non-empty | Returns input |
| `maskString()` | Any input | Returns masked string |
| `saveConfig()` | Valid config | Writes files correctly |
| `saveConfig()` | YAML error | Returns error |

## 5. Mocking Requirements

### 5.1 Telegram Bot Mock

Mock `*tgbot.Bot` to verify:
- `SendMessage` calls with correct parameters
- `SendChatAction` for typing indicator
- Handler registration calls

### 5.2 LLM Provider Mock

Mock `llm.Provider` interface:
- Return custom responses
- Return specific enabled states
- Return configurable errors

### 5.3 Session Manager Mock

Mock `session.Manager` interface:
- Return custom session data
- Return configurable errors

## 6. Test Patterns to Follow

Use patterns from `internal/config/config_test.go`:
- Use `t.TempDir()` for temp files
- Use table-driven tests where appropriate
- Clear environment variables in `TestMain` or each test
- Test error cases explicitly
- Name tests: `Test<Function>_<Scenario>`

## 7. Coverage Goals

| Package | Target |
|---------|--------|
| `internal/llm` | 70%+ |
| `internal/session` | 80%+ |
| `internal/bot` | 70%+ |
| `cmd/setup` | 50%+ |

## 8. Exclusions

The following are explicitly NOT tested in unit tests:
- `cmd/bot/main.go` - Entry point with side effects
- Live API calls - Integration test territory
- Network timeout actual values - Configuration dependent