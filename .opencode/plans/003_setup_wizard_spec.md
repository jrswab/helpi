# Setup Wizard Specification

## Overview

The Setup Wizard is a CLI tool (`go run ./cmd/setup`) that initializes or updates the Helpi bot configuration. It handles both initial setup and subsequent configuration updates by detecting existing configuration files and pre-filling values.

## File Locations

- **Executable**: `cmd/setup/main.go`
- **Config Output**: `config.yaml` (in working directory)
- **Secrets Output**: `.env` (in working directory)

## Execution Modes

The wizard operates in two modes:
1. **Initial Setup**: No existing configuration files detected
2. **Update Mode**: Existing `config.yaml` and/or `.env` detected

## Input Sources

| Source | Priority | Description |
|--------|----------|-------------|
| stdin | Primary | User input for all prompts |
| config.yaml | Secondary | Pre-fill non-sensitive defaults |
| .env | Secondary | Pre-fill masked API keys |

## Workflow Steps

### Step 1: Detect Existing Configuration

**Action**: Check for existing configuration files in the following order:
1. Current working directory
2. Directory containing the executable binary

**Behavior**:
- If `config.yaml` exists: Load and display current values as defaults
- If `.env` exists: Load API keys but mask them from display (`[SET]` placeholder)
- If neither exists: Proceed as initial setup with empty defaults

**Edge Cases**:
- Config file exists but is malformed: Abort with error message showing the issue
- Config file exists but is empty: Treat as non-existent
- Both config.yaml and .env exist but are inconsistent: Prefer config.yaml for enabled/disabled states, .env for API keys

### Step 2: Telegram Bot Token

**Prompt Text**:
```
Telegram Bot Token (required):
```

**Behavior**:
- Show current value if exists (masked: `****1234`)
- Empty input preserves existing value
- Non-empty input replaces existing value

**Validation**:
- Must be non-empty if no existing value
- Must be non-empty if existing value is being replaced
- If user enters empty and no existing value exists: Show error, re-prompt

**Edge Cases**:
- Invalid token format (not matching Bot API pattern): Show warning but allow (tokens validated at runtime)
- Token starts with valid prefix but is incomplete: Accept with warning

### Step 3: Provider Configuration

**Providers Available**:
| Provider | Default Model (if enabled) |
|----------|---------------------------|
| OpenAI | gpt-4o |
| Anthropic | claude-3-5-sonnet-20241022 |
| OpenRouter | openai/gpt-4o |
| OpenCode | opencode/big-pickle |
| Ollama | llama3.2 |

**Prompt for each provider**:
```
Enable [Provider]? (y/n) [current state]:
```

**Behavior**:
- Default input (empty) preserves current enabled/disabled state
- `y`, `Y`, `yes`, `YES` enables the provider
- `n`, `N`, `no`, `NO` disables the provider
- Any other input: Show error, re-prompt

**Edge Cases**:
- Provider currently enabled but .env has no API key: Warn user that provider won't work
- Provider currently enabled but config has no default model: Use provider's default model

### Step 4: Per-Enabled Provider Configuration

For each provider marked as enabled in Step 3:

#### 4a. API Key

**Prompt Text**:
```
[Provider] API Key:
```

**Behavior**:
- If existing `.env` has API key: Display `[SET]` as placeholder
- Empty input preserves existing API key if set
- Non-empty input replaces existing API key
- If no existing key and user enters empty: Show error, re-prompt

**Masking Rules**:
- When displaying current key: Always show `[SET]` regardless of actual key value
- When user types: Input is not echoed to terminal

**Edge Cases**:
- User enters `[SET]` as literal text: Treat as empty (preserve existing)
- User clears existing key: Not supported; must enter new key or leave as-is
- Provider disabled but has existing key in .env: Keep key in .env (don't remove)

#### 4b. Default Model

**Prompt Text**:
```
[Provider] default model [default]:
```

Where `[default]` is the provider's built-in default if no existing config.

**Behavior**:
- Empty input uses the displayed default
- Non-empty input uses entered model string

**Validation**:
- Must be non-empty string
- No format validation (model names vary by provider)

**Edge Cases**:
- Model name contains spaces: Allow (some providers use this)
- User enters model not available to them: Allow (will fail at runtime with clear error)

### Step 5: Allowed Users

**Prompt Text**:
```
Allowed Telegram User IDs (comma-separated):
```

**Behavior**:
- Show current list if exists (comma-separated, e.g., `123456789, 987654321`)
- Empty input preserves existing list
- Non-empty input: Parse comma-separated IDs, trim whitespace

**Validation**:
- Each ID must be numeric (Telegram user IDs are integers)
- At least one ID must be configured (enforced at runtime via config validation)

**Edge Cases**:
- User enters non-numeric value: Show error "Invalid user ID: [value]. Must be numeric.", re-prompt
- User enters duplicate IDs: Remove duplicates automatically
- User enters negative numbers: Allow (some legacy Telegram IDs were negative)
- Whitespace around IDs: Trim automatically
- Extra commas (e.g., `123, , 456`): Treat as empty entries, ignore

### Step 6: Memory Configuration

#### 6a. Session Path

**Prompt Text**:
```
Session storage path [default: ./data/sessions]:
```

**Behavior**:
- Empty input uses default `./data/sessions`
- Non-empty input uses entered path

**Validation**:
- Path can be relative or absolute
- No validation that path exists or is writable (deferred to runtime)

**Edge Cases**:
- Path contains tilde (~): Do NOT expand; treat as literal
- Path contains spaces: Allow (quote if needed when used)

#### 6b. Max Messages

**Prompt Text**:
```
Max messages per conversation (0 to retain all) [default: 50]:
```

**Behavior**:
- Empty input uses default `50`
- Non-empty input: Parse as integer

**Validation**:
- Must be integer >= 0
- 0 means no limit (retain all messages)

**Edge Cases**:
- User enters negative: Show error, re-prompt
- User enters non-numeric: Show error, re-prompt
- User enters decimal: Show error, re-prompt

### Step 7: Save Configuration

**Action**: Write two files:

#### config.yaml
- Write to current working directory
- Include all non-sensitive configuration
- Use YAML format with 2-space indentation
- Include comments for clarity

#### .env
- Write to current working directory
- Include all API keys
- Do NOT include any non-API-key values
- Overwrite existing file entirely

**Edge Cases**:
- Config file cannot be written: Show error with reason, offer retry
- Partial write failure: Attempt to clean up partial file, show error
- .env already exists with other variables (not from helpi): Preserve other variables, only update helpi-related keys

## Output

On successful completion:
```
✓ Configuration saved to config.yaml
✓ Secrets saved to .env

Run the bot with: go run ./cmd/bot
```

On failure:
```
✗ Error: [description]
```

No colors are used in output (terminal compatibility).

## Non-Interactive Mode

Not supported. Setup wizard always runs interactively.

## Keyboard Interrupts

- Ctrl+C during any prompt: Abort without saving, show message "Setup cancelled."
- No partial configuration is saved on abort.

## Summary of Default Values

| Setting | Default |
|---------|---------|
| Ollama base URL | http://localhost:11434 |
| Session path | ./data/sessions |
| Max messages | 50 |
| OpenAI default model | gpt-4o |
| Anthropic default model | claude-3-5-sonnet-20241022 |
| OpenRouter default model | openai/gpt-4o |
| OpenCode default model | opencode/big-pickle |
| Ollama default model | llama3.2 |