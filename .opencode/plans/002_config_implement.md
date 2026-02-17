# Config Struct and Loader Implementation

Based on: `001_config_struct_loader_spec.md`

## Implementation Checklist

- [x] Create configuration data structures in `internal/config/config.go`
  - Define `Config` struct with Telegram, AllowedUsers, Providers, Memory fields
  - Define `TelegramConfig` struct with Token field
  - Define `ProviderConfig` struct with Enabled and DefaultModel fields
  - Define `ProvidersConfig` struct with OpenAI, Anthropic, OpenRouter, OpenCode, Ollama fields
  - Define `MemoryConfig` struct with Path and MaxMessages fields
  - Add yaml tags to all struct fields

- [x] Create config loader in `internal/config/loader.go`
  - Implement `Load()` function that loads and returns validated Config
  - Implement `findConfigDir()` to search current directory and executable directory
  - Implement `loadYAML()` to parse config.yaml
  - Implement `loadEnv()` to parse .env file

- [x] Add validation functions in `internal/config/loader.go`
  - Implement `validateConfig()` to validate all fields
  - Validate telegram.token is non-empty
  - Validate allowed_users is present (can be empty array)
  - Validate each enabled provider has non-empty default_model
  - Validate memory.max_messages is >= 1

- [x] Add environment variable validation in `internal/config/loader.go`
  - Check required API keys based on enabled providers
  - OPENAI_API_KEY required when openai.enabled is true
  - ANTHROPIC_API_KEY required when anthropic.enabled is true
  - OPENROUTER_API_KEY required when openrouter.enabled is true
  - OPENCODE_API_KEY required when opencode.enabled is true
  - OLLAMA_BASE_URL is optional (default to http://localhost:11434 if missing)

- [x] Create .env template file in project root
  - Create `.env.example` with all required variables
  - Document each variable with comments
  - Include TELEGRAM_BOT_TOKEN, OPENAI_API_KEY, ANTHROPIC_API_KEY, OPENROUTER_API_KEY, OPENCODE_API_KEY, OLLAMA_BASE_URL

- [x] Add error types and error messages in `internal/config/loader.go`
  - Define custom error type (e.g., ConfigError) with descriptive messages
  - Include file path in error messages
  - Include field name and reason for validation failures
  - Handle YAML parse errors with line numbers
  - Handle missing/malformed .env variables

- [x] Write unit tests in `internal/config/config_test.go`
  - Test successful loading with valid config
  - Test missing required fields returns error
  - Test missing API keys for enabled providers returns error
  - Test empty allowed_users array is valid
  - Test all providers disabled is valid
  - Test unknown YAML fields are ignored (lenient parsing)
  - Test whitespace trimming on string values