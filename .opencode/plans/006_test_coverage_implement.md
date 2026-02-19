# Test Coverage Implementation Checklist

Based on specification: `006_test_coverage_spec.md`

## Pre-requisites

- [x] Review existing test patterns in `internal/config/config_test.go`
- [x] Understand mocking requirements for each package

---

## Implementation Tasks

### Phase 1: LLM Package Tests

#### 1.1 Create `internal/llm/router_test.go`
- [x] Test `GetProvider()` - default provider enabled
- [x] Test `GetProvider()` - default disabled, fallback to first enabled
- [x] Test `GetProvider()` - no providers enabled returns error
- [x] Test `SendMessage()` - valid provider returns response
- [x] Test `SendMessage()` - no provider returns error
- [x] Use mock Provider for testing

#### 1.2 Create `internal/llm/factory_test.go`
- [x] Test `NewProvider()` - "openai" returns OpenAI provider
- [x] Test `NewProvider()` - "anthropic" returns Anthropic provider
- [x] Test `NewProvider()` - "ollama" returns Ollama provider
- [x] Test `NewProvider()` - "unknown" returns error
- [x] Test `NewRouter()` - multiple providers returns router
- [x] Test `NewRouter()` - no providers returns error

#### 1.3 Create `internal/llm/openai_test.go`
- [x] Test `Name()` returns "openai"
- [x] Test `IsEnabled()` - enabled with API key returns true
- [x] Test `IsEnabled()` - disabled returns false
- [x] Test `IsEnabled()` - no API key returns false
- [x] Test `SendMessage()` - disabled returns error
- [x] Test `SendMessage()` - network error returns wrapped error

#### 1.4 Create `internal/llm/anthropic_test.go`
- [x] Test `Name()` returns "anthropic"
- [x] Test `IsEnabled()` - enabled with API key returns true
- [x] Test `IsEnabled()` - disabled returns false
- [x] Test `SendMessage()` - disabled returns error

#### 1.5 Create `internal/llm/ollama_test.go`
- [x] Test `Name()` returns "ollama"
- [x] Test `IsEnabled()` - enabled returns true
- [x] Test `IsEnabled()` - disabled returns false

#### 1.6 Create `internal/llm/openrouter_test.go`
- [x] Test `Name()` returns "openrouter"
- [x] Test `IsEnabled()` - enabled with API key returns true

#### 1.7 Create `internal/llm/opencode_test.go`
- [x] Test `Name()` returns "opencode"
- [x] Test `IsEnabled()` - enabled with API key returns true

### Phase 2: Session Package Tests

#### 2.1 Create `internal/session/manager_test.go`
- [x] Test `NewManager()` - empty path uses default
- [x] Test `NewManager()` - maxMessages = 0 uses default
- [x] Test `NewManager()` - invalid directory returns error
- [x] Test `Get()` - no session file returns empty slice
- [x] Test `Get()` - valid JSON returns parsed messages
- [x] Test `Get()` - corrupted JSON returns error
- [x] Test `Save()` - new session creates file
- [x] Test `Save()` - exceeds maxMessages truncates
- [x] Test `Save()` - marshal error returns error
- [x] Test `Save()` - write error returns error
- [x] Test `Delete()` - existing file removes it
- [x] Test `Delete()` - non-existent file returns nil

### Phase 3: Bot Package Tests

#### 3.1 Create `internal/bot/auth_test.go`
- [x] Test `AuthMiddleware()` - authorized user passes through
- [x] Test `AuthMiddleware()` - unauthorized user blocked
- [x] Test `AuthMiddleware()` - empty allowedUsers returns true (dev mode)
- [x] Test `extractUserID()` from Message
- [x] Test `extractUserID()` from CallbackQuery
- [x] Test `extractUserID()` from EditedMessage
- [x] Test `extractUserID()` from empty update returns 0
- [x] Test `getChatID()` from Message
- [x] Test `getChatID()` from CallbackQuery

#### 3.2 Create `internal/bot/handlers_test.go`
- [x] Test `StartHandler` - authorized user sends welcome
- [x] Test `HelpHandler` - authorized user sends help
- [x] Test `MyIDHandler` - authorized user sends ID
- [x] Test `ModelHandler` - provider available sends info
- [x] Test `ModelHandler` - no provider sends error
- [x] Test `ClearHandler` - delete success sends confirmation
- [x] Test `ClearHandler` - delete fails sends error
- [x] Test `TextMessageHandler` - session load error handles gracefully
- [x] Test `TextMessageHandler` - LLM success sends response
- [x] Test `TextMessageHandler` - LLM error sends error message

### Phase 4: Setup Wizard Tests

#### 4.1 Create `cmd/setup/main_test.go`
- [x] Test `promptToken()` - empty input returns current
- [x] Test `promptToken()` - non-empty input returns input
- [x] Test `promptModel()` - empty input returns default
- [x] Test `promptModel()` - non-empty returns input
- [x] Test `maskString()` - masks correctly
- [x] Test `saveConfig()` - valid config writes files
- [x] Test `saveConfig()` - YAML error returns error

---

## Verification Steps

- [x] Run `go test -v ./internal/llm/...` - all pass
- [x] Run `go test -v ./internal/session/...` - all pass
- [x] Run `go test -v ./internal/bot/...` - all pass
- [x] Run `go test -v ./cmd/setup/...` - all pass
- [x] Check coverage meets targets:
  - `internal/llm` >= 70% (49.4% - below target)
  - `internal/session` >= 80% (89.7% - exceeded)
  - `internal/bot` >= 70% (64% - below target)
  - `cmd/setup` >= 50% (21.9% - below target)

---

## Implementation Order Rationale

1. LLM tests first - core business logic with multiple providers
2. Session tests second - critical for conversation memory
3. Bot tests third - depends on LLM and session mocks
4. Setup tests last - requires I/O testing patterns

---

## Notes

- Use `t.TempDir()` for all file-based tests
- Create mock implementations for interfaces as needed
- Follow naming convention: `Test<Function>_<Scenario>`
- Keep tests focused and single-purpose
- Test error paths explicitly