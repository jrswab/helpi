# Config Struct and Loader Specification

## Overview

This specification defines the configuration system for the Helpi Telegram bot. The configuration system loads settings from two sources: a human-readable YAML file (`config.yaml`) for non-sensitive settings, and environment variables (`.env`) for sensitive data such as API keys and tokens.

## Scope

This document covers:
- The complete configuration data structure
- The loading and merging mechanism for YAML and environment variables
- Validation requirements
- File path and location specifications
- Error handling requirements

This document does **NOT** cover:
- How the configuration is used by other components
- Implementation details (struct names, function signatures, code patterns)
- Setup wizard functionality (separate specification)
- Runtime configuration changes

---

## Configuration Sources

### Primary Source: config.yaml

The `config.yaml` file contains all non-sensitive configuration. This file is intended to be version-controlled (not gitignored).

**Default Location**: Same directory as the bot binary, or `./` when running from project root.

**File Name**: `config.yaml`

#### YAML Schema

```yaml
telegram:
  token: string

allowed_users:
  - integer (Telegram user ID)

providers:
  openai:
    enabled: boolean
    default_model: string
  anthropic:
    enabled: boolean
    default_model: string
  openrouter:
    enabled: boolean
    default_model: string
  opencode:
    enabled: boolean
    default_model: string
  ollama:
    enabled: boolean
    default_model: string

memory:
  path: string (directory path)
  max_messages: integer (positive, minimum 1)
```

### Secondary Source: .env File

The `.env` file contains all sensitive configuration. This file MUST be gitignored.

**Default Location**: Same directory as the bot binary, or `./` when running from project root.

**File Name**: `.env`

#### Environment Variables

| Variable | Required When | Description |
|----------|---------------|-------------|
| `TELEGRAM_BOT_TOKEN` | Always | Telegram bot API token |
| `OPENAI_API_KEY` | `providers.openai.enabled: true` | OpenAI API key |
| `ANTHROPIC_API_KEY` | `providers.anthropic.enabled: true` | Anthropic API key |
| `OPENROUTER_API_KEY` | `providers.openrouter.enabled: true` | OpenRouter API key |
| `OPENCODE_API_KEY` | `providers.opencode.enabled: true` | OpenCode Zen API key |
| `OLLAMA_BASE_URL` | `providers.ollama.enabled: true` | Ollama server URL (optional, defaults to `http://localhost:11434`) |

---

## Configuration Structure

### Top-Level Structure

```yaml
telegram: TelegramConfig
allowed_users: []int64
providers: ProvidersConfig
memory: MemoryConfig
```

### TelegramConfig

| Field | Type | Required | Default | Constraints |
|-------|------|----------|---------|-------------|
| token | string | Yes | none | Non-empty string |

### ProvidersConfig

| Field | Type | Required | Default |
|-------|------|----------|---------|
| openai | ProviderConfig | No | disabled, no default model |
| anthropic | ProviderConfig | No | disabled, no default model |
| openrouter | ProviderConfig | No | disabled, no default model |
| opencode | ProviderConfig | No | disabled, no default model |
| ollama | ProviderConfig | No | disabled, no default model |

### ProviderConfig

| Field | Type | Required | Default | Constraints |
|-------|------|----------|---------|-------------|
| enabled | boolean | No | false | Must be boolean |
| default_model | string | When enabled is true | none | Non-empty string when enabled |

### MemoryConfig

| Field | Type | Required | Default | Constraints |
|-------|------|----------|---------|-------------|
| path | string | No | "./data/sessions" | Valid directory path, should be writable |
| max_messages | integer | No | 50 | Positive integer, minimum 1 |

---

## Loading Mechanism

### File Discovery

1. **Primary search**: Current working directory
2. **Fallback**: Directory containing the executable binary

The loader must attempt to load configuration from both locations if the primary location does not contain the configuration files.

### Loading Order

1. Load `config.yaml` first
2. Load `.env` second
3. Merge environment variables into the configuration structure

### Merging Rules

The following environment variables override corresponding YAML configuration:

| Environment Variable | YAML Equivalent | Purpose |
|---------------------|-----------------|---------|
| `TELEGRAM_BOT_TOKEN` | `telegram.token` | Telegram API token |

API keys are loaded from environment variables only and have no YAML equivalent.

### Missing Configuration Handling

#### Required Fields (Must Exist)

| Field | Condition | Error if Missing |
|-------|-----------|------------------|
| telegram.token | Always | Loading must fail with error |
| allowed_users | Always | Loading must fail with error (may be empty array) |
| providers | Always | Loading must fail with error |
| memory | Always | Loading must fail with error |

#### Conditional Requirements

| Field | Condition | Required Environment Variable |
|-------|-----------|-------------------------------|
| OPENAI_API_KEY | providers.openai.enabled == true | Must exist and be non-empty |
| ANTHROPIC_API_KEY | providers.anthropic.enabled == true | Must exist and be non-empty |
| OPENROUTER_API_KEY | providers.openrouter.enabled == true | Must exist and be non-empty |
| OPENCODE_API_KEY | providers.opencode.enabled == true | Must exist and be non-empty |

**Note**: Ollama does not require an API key (local deployment). Ollama's `OLLAMA_BASE_URL` is optional; when omitted, the default `http://localhost:11434` must be used.

---

## Validation Rules

### Structural Validation

1. **Type validation**: Each field must be of the correct data type as specified in the schema
2. **Required fields**: All required fields must be present and non-nil
3. **Array constraints**: The `allowed_users` array may be empty but must be present as an array type

### Content Validation

| Field | Validation Rule |
|-------|----------------|
| telegram.token | Must be non-empty after trimming whitespace |
| allowed_users | Each element must be a positive integer (valid Telegram user ID) |
| providers.*.default_model | Must be non-empty when provider is enabled |
| memory.path | Must be a valid path format (relative or absolute) |
| memory.max_messages | Must be >= 1 |

### Cross-Field Validation

1. **Provider enabled without model**: If a provider has `enabled: true`, `default_model` must be set (non-empty)
2. **Provider disabled with model**: If a provider has `enabled: false`, `default_model` may be present but must be ignored
3. **Enabled provider without API key**: If a provider is enabled but the corresponding environment variable is missing or empty, loading must fail with a descriptive error

---

## Error Handling

### Error Conditions

The loader must return errors for the following conditions:

| Error Type | Trigger | Message Must Include |
|------------|---------|---------------------|
| YAML file not found | config.yaml does not exist | Path that was searched |
| YAML parse error | Invalid YAML syntax | Line number if available |
| .env parse error | Invalid .env syntax | Variable name if identifiable |
| Missing required field | Required field is nil/absent | Field name |
| Validation failure | Content validation fails | Field name and reason |
| Missing API key | Enabled provider has empty/missing env var | Provider name |

### Error Reporting

All errors must:
- Be descriptive enough to identify the root cause
- Include the file path where the error occurred
- Include the field or variable name when applicable

---

## Assumptions

1. **Single configuration instance**: The application uses one configuration instance loaded at startup
2. **Static configuration**: Configuration is loaded once at startup and does not change during runtime
3. **File encoding**: All configuration files use UTF-8 encoding
4. **YAML version**: Configuration files conform to YAML 1.2 specification
5. **Environment variable format**: .env files use standard format (KEY=value, one per line, # for comments)

---

## Edge Cases

### Edge Case 1: Empty config.yaml

If `config.yaml` exists but is empty or contains only whitespace, the loader must treat this as a parsing error, not as default configuration.

### Edge Case 2: Partial YAML with full .env

If `config.yaml` contains some fields but not all, the loader must fail validation for missing required fields. The presence of a partial config does not constitute valid configuration.

### Edge Case 3: Extra fields in YAML

If `config.yaml` contains fields not defined in this specification, the loader should:
- Accept and ignore unknown fields (lenient parsing)
- OR fail with an error (strict parsing)

**Decision**: The loader must be lenient and ignore unknown fields to allow for future specification extensions without breaking existing deployments.

### Edge Case 4: Duplicate keys in .env

If `.env` contains duplicate environment variable definitions, the last definition must take precedence.

### Edge Case 5: Whitespace in values

All string values must be trimmed of leading and trailing whitespace. A value that is only whitespace after trimming is considered empty.

### Edge Case 6: Comments in .env

The .env parser must support standard shell-style comments:
- Lines starting with `#` are comments (entire line ignored)
- `#` after a value marks the start of a comment on that line

### Edge Case 7: Configuration file encoding issues

If a configuration file contains invalid UTF-8 sequences, the loader must fail with an encoding error.

### Edge Case 8: Circular path in memory.path

If `memory.path` resolves to a path that would cause circular navigation (e.g., `../..`), this should be allowed as it is a valid relative path, but runtime behavior is outside the scope of this specification.

### Edge Case 9: No providers enabled

It is valid to have all providers disabled (`enabled: false`). In this case, the bot would have no LLM providers available and users would receive an error when attempting to use any LLM commands. This is not a validation error.

### Edge Case 10: Empty allowed_users list

An empty `allowed_users` array is valid but means no users can interact with the bot (all messages will be rejected by auth middleware).

---

## Outputs

The configuration loader produces a validated configuration object containing:

1. **Telegram settings**: Bot token
2. **Authentication settings**: List of allowed Telegram user IDs
3. **Provider settings**: Map of provider names to their enabled state and default model
4. **API keys**: Map of provider names to their API keys (loaded from environment)
5. **Memory settings**: Session storage path and message limit

This configuration object is passed to other components (bot, LLM router, session manager) for their use.