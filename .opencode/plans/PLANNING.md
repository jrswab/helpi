# Helpi - Telegram LLM Bot

## Project Overview

Multi-provider Telegram bot that connects to hosted LLMs (OpenAI, Anthropic, Ollama, OpenRouter, OpenCode Zen) with file-based conversation memory and user whitelist authentication.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Telegram Bot (Go)                       │
├─────────────────────────────────────────────────────────────┤
│  Commands        │  Message Handler  │  Auth Middleware     │
│  /start          │  (go-telegram/bot)│  (user whitelist)   │
│  /model          │                   │                      │
│  /clear          │                   │                      │
│  /myid           │                   │                      │
└──────┬───────────┴────────┬──────────┴──────────┬───────────┘
       │                    │                      │
       ▼                    ▼                      ▼
┌─────────────────────────────────────────────────────────────┐
│                    LLM Router (Interface)                   │
└──────┬──────┬───────┬───────┬────────────┬─────────────────┘
       │      │       │       │            │
   ┌───┴───┐  │   ┌───┴───┐   │   ┌────────┴────────┐
   │OpenAI │  │   │Anthropic   │   │    Ollama       │
   │(go-   │  │   │(official)  │   │  (native http)  │
   │openai)│  │   │            │   │                 │
   └───┬───┘  │   └─────┬─────┘   └─────────────────┘
       │      │         │
   ┌───┴───┐  │   ┌─────┴─────┐
   │OpenRouter   │   │OpenCode Zen│
   │(go-openai   │   │(native http)│
   │ + base URL) │   │(OpenAI-compat)
   └─────────┘  └───────────────┘
```

## Project Structure

```
helpi/
├── cmd/
│   ├── bot/main.go              # Bot entry point
│   └── setup/main.go            # CLI setup wizard
├── internal/
│   ├── bot/
│   │   ├── handlers.go          # Message/command handlers
│   │   └── auth.go              # User whitelist middleware
│   ├── llm/
│   │   ├── router.go            # Provider interface
│   │   ├── openai.go            # OpenAI provider
│   │   ├── anthropic.go         # Anthropic provider
│   │   ├── ollama.go            # Ollama provider
│   │   ├── openrouter.go        # OpenRouter provider
│   │   └── opencode.go          # OpenCode Zen provider
│   ├── session/
│   │   └── manager.go           # File-based conversation memory
│   └── config/
│       ├── config.go            # Configuration struct
│       └── loader.go            # Config loading from YAML + .env
├── config.yaml                  # Bot config (non-sensitive)
├── .env                         # API keys (gitignored)
└── go.mod
```

## Tech Stack

- **Language**: Go 1.22+
- **Telegram**: [go-telegram/bot](https://github.com/go-telegram/bot)
- **OpenAI**: [openai/openai-go/v3](https://github.com/openai/openai-go) (official)
- **Anthropic**: [anthropics/anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go) (official)
- **Ollama**: Native HTTP (OpenAI-compatible: `http://localhost:11434/v1`)
- **OpenRouter**: `go-openai` with custom base URL (`https://openrouter.ai/api/v1`)
- **OpenCode Zen**: Native HTTP (OpenAI-compatible: `https://opencode.ai/zen/v1/responses`)
- **Config**: [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml)

## Setup Wizard Flow (`go run ./cmd/setup`)

```
┌─────────────────────────────────────────────────────────────┐
│ 1. DETECT EXISTING CONFIG                                   │
│    → If config.yaml exists: load and pre-fill defaults      │
│    → If .env exists: load and mask existing API keys        │
├─────────────────────────────────────────────────────────────┤
│ 2. Telegram Bot Token                                       │
│    → Show current if exists, prompt for new if needed       │
├─────────────────────────────────────────────────────────────┤
│ 3. Enable Providers (show current state, toggle y/n)        │
│    → OpenAI, Anthropic, Ollama, OpenRouter, OpenCode        │
├─────────────────────────────────────────────────────────────┤
│ 4. Per Enabled Provider:                                    │
│    → API key (masked input, show [SET] if already exists)   │
│    → Default model                                          │
├─────────────────────────────────────────────────────────────┤
│ 5. Allowed Users                                            │
│    → Show current list                                      │
│    → Option to add/remove IDs                               │
├─────────────────────────────────────────────────────────────┤
│ 6. Save to config.yaml + .env                               │
└─────────────────────────────────────────────────────────────┘
```

## Configuration Files

### config.yaml (non-sensitive)
```yaml
telegram:
  token: "BOT_TOKEN"

allowed_users:
  - 123456789
  - 987654321

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

### .env (API keys, gitignored)
```
TELEGRAM_BOT_TOKEN=...
OPENAI_API_KEY=...
ANTHROPIC_API_KEY=...
OPENROUTER_API_KEY=...
OPENCODE_API_KEY=...
```

## Commands

| Command | Description |
|---------|-------------|
| `/start` | Welcome + list commands |
| `/myid` | Reply with user's Telegram ID |
| `/model` | Show available models/providers |
| `/clear` | Clear conversation history |
| `<text>` | Send message to LLM |

## User Whitelist Flow

1. Run setup → configure bot + providers
2. Start bot → user sends `/myid`
3. Bot replies with their ID
4. Run setup again → add ID to allowed_users
5. Restart bot

## Implementation Order

1. Initialize Go module with dependencies
2. Config struct + loader (YAML + .env)
3. Setup wizard (handles both new + update)
4. LLM provider interface + implementations
5. Telegram bot + handlers + auth middleware
6. Session manager (file-based JSON)
7. Build and test