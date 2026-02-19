# Telegram Bot Specification

## 1. Overview

This document specifies the Telegram bot layer for the Helpi application. The bot layer handles all Telegram user interactions, enforces user authentication via whitelist, processes commands, routes messages to the LLM layer, and manages conversation history.

## 2. Scope

### 2.1 In Scope

- Telegram bot initialization and event polling
- User authentication via whitelist (auth middleware)
- Command handlers: /start, /help, /myid, /model, /clear
- Text message handling with LLM response generation
- Integration with LLM router for message processing
- Integration with session manager for conversation history

### 2.2 Out of Scope

- Webhook-based bot deployment (polling only)
- Inline query handling
- Callback query handling (keyboard buttons)
- Media message handling (images, documents, voice)
- Group chat handling (single user mode only)
- Rate limiting or flood control

## 3. Design Goals

1. **Security First**: All messages must pass through auth middleware before processing
2. **Idempotent Commands**: Commands can be repeated safely without side effects
3. **Graceful Degradation**: Bot remains functional even if LLM is temporarily unavailable
4. **User-Friendly Errors**: All error messages are actionable and user-friendly

## 4. Architecture

### 4.1 Component Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Telegram Bot (Go)                         │
├─────────────────────────────────────────────────────────────┤
│  Entry Point         │  Handlers           │  Middleware    │
│  cmd/bot/main.go     │  handlers.go        │  auth.go       │
│                      │                     │                │
│  - Config loading    │  - /start           │  - User ID     │
│  - Bot init          │  - /help            │    validation  │
│  - Event polling     │  - /myid            │  - Whitelist   │
│  - Handler reg       │  - /model           │    checking    │
│                      │  - /clear           │                │
│                      │  - text messages    │                │
└──────────────────────┴─────────────────────┴────────────────┘
```

### 4.2 Dependencies

The bot layer depends on:
- `internal/config` - Configuration loading and access
- `internal/llm` - LLM router interface for message processing
- `internal/session` - Session manager for conversation history
- `github.com/go-telegram/bot` - Telegram Bot API library

## 5. File Structure

```
cmd/bot/
└── main.go           # Bot entry point

internal/bot/
├── handlers.go       # Command and message handlers
├── auth.go           # Authentication middleware
└── bot.go            # Bot initialization helpers
```

## 6. Bot Entry Point (cmd/bot/main.go)

### 6.1 Initialization Sequence

The main function MUST perform the following steps in order:

1. Load configuration from config.yaml and .env using `config.Load()`
2. Validate that telegram bot token is configured (exit with error if missing)
3. Validate that at least one LLM provider is enabled (exit with error if none)
4. Initialize LLM router using `llm.NewRouter(config)`
5. Initialize session manager using `session.NewManager(config.Memory)`
6. Initialize Telegram bot using `bot.New()`
7. Register auth middleware
8. Register all command handlers
9. Register message handler
10. Start polling using `bot.Start()`

### 6.2 Configuration Requirements

The following configuration values MUST be present:

| Config Field | Source | Validation |
|--------------|--------|------------|
| `telegram.token` | config.yaml | Must be non-empty string |
| `allowed_users` | config.yaml | Must be non-empty slice (can be empty during initial setup) |
| `providers.*.enabled` | config.yaml | At least one must be true |

### 6.3 Logging

- Bot MUST log startup messages including "Starting Helpi bot..."
- Bot MUST log the number of allowed users on startup
- Bot MUST log when polling starts

## 7. Authentication Middleware (internal/bot/auth.go)

### 7.1 Purpose

The auth middleware ensures that only whitelisted users can interact with the bot. It is applied to all incoming updates before any handler is executed.

### 7.2 Middleware Implementation

```go
func AuthMiddleware(config *config.Config) HandlerFunc
```

The middleware function MUST:

1. Extract the user ID from the incoming update
2. Check if the user ID exists in `config.AllowedUsers`
3. If user is not whitelisted:
   - Send a private message to the user: "Access denied. Your Telegram ID is not authorized to use this bot."
   - Log the unauthorized access attempt with user ID
   - Return without further processing
4. If user is whitelisted:
   - Pass the update to the next handler

### 7.3 User ID Extraction

| Update Type | User ID Source |
|-------------|----------------|
| Message | `update.Message.From.ID` |
| CallbackQuery | `update.CallbackQuery.From.ID` |
| EditedMessage | `update.EditedMessage.From.ID` |

### 7.4 Edge Cases

- If update does not contain user information: Log error, skip processing
- If `allowed_users` is empty: Allow all users (development mode warning logged)

## 8. Command Handlers (internal/bot/handlers.go)

### 8.1 /start Command

**Trigger**: User sends `/start`

**Behavior**:
1. Send welcome message with bot description
2. List all available commands with descriptions

**Response Format**:
```
👋 Welcome to Helpi!

I'm a multi-provider AI assistant. I can help you with questions, conversations, and more.

Available commands:
/help - Show this help message
/model - Show available models and providers
/myid - Get your Telegram ID
/clear - Clear conversation history
/question - Ask me anything (just send a message)
```

### 8.2 /help Command

**Trigger**: User sends `/help`

**Behavior**:
1. Send detailed help message

**Response Format**:
```
📖 Helpi Help

Commands:
/start - Welcome message
/help - Show this help
/model - Show current model and available providers
/myid - Get your Telegram ID (for adding to whitelist)
/clear - Clear conversation history

Usage:
- Just send me a message to chat with the AI
- Use /clear to start a new conversation
- Use /model to check which AI model is active
```

### 8.3 /myid Command

**Trigger**: User sends `/myid`

**Behavior**:
1. Extract the user's Telegram ID
2. Send a message with the user's ID

**Response Format**:
```
Your Telegram ID: {user_id}
```

**Use Case**: Users run this command to get their ID, then provide it to the bot admin for adding to the whitelist.

### 8.4 /model Command

**Trigger**: User sends `/model`

**Behavior**:
1. Get the current active provider from the router
2. List all configured providers with their enabled/disabled status

**Response Format**:
```
🤖 Current Model

Active: {provider_name} ({model_name})

Available Providers:
✓ OpenAI: {model_name} (enabled)
✗ Anthropic: (disabled)
✓ OpenRouter: {model_name} (enabled)
✓ OpenCode: {model_name} (enabled)
✗ Ollama: (disabled)
```

### 8.5 /clear Command

**Trigger**: User sends `/clear`

**Behavior**:
1. Clear the user's session data using session manager
2. Send confirmation message

**Response Format**:
```
🗑️ Conversation cleared! Starting fresh.
```

**Edge Cases**:
- If session file does not exist: Still send success message (idempotent)
- If session deletion fails: Send error message with details

## 9. Message Handler

### 9.1 Text Message Processing

**Trigger**: User sends any text message that is not a command

**Behavior**:
1. Build message history from session manager
2. Add user message to history
3. Send messages to LLM router
4. Add assistant response to history
5. Save session using session manager
6. Send response to user

### 9.2 Session Management

Each user has an independent conversation history stored in the session manager. The session manager handles:
- Loading existing conversation history
- Saving updated conversation history
- Managing session file naming and location

### 9.3 Message Flow

```
User Message → Load Session → Add to History → LLM Router → 
Save Session → Send Response → Done
```

### 9.4 LLM Response Handling

| Scenario | Behavior |
|----------|----------|
| Success | Send response text to user |
| No provider enabled | Send: "No AI provider configured. Please contact admin." |
| API error | Send: "Sorry, I encountered an error: {error_message}. Please try again." |
| Timeout | Send: "Request timed out. Please try again." |
| Empty response | Send: "I didn't get a response. Please try again." |

### 9.5 Typing Indicator

The bot SHOULD show a typing indicator while waiting for the LLM response to improve user experience.

## 10. Error Handling

### 10.1 Logging Requirements

All errors MUST be logged with:
- Timestamp
- User ID (if available)
- Error context
- Error message

### 10.2 User-Facing Errors

| Error Type | User Message |
|------------|--------------|
| Auth failure | "Access denied. Your Telegram ID is not authorized to use this bot." |
| LLM API error | "Sorry, I encountered an error: {brief_error}. Please try again." |
| Session error | "Failed to save conversation. Your messages may not be remembered." |
| Config missing | "Bot configuration error. Please contact admin." |

### 10.3 No Panic Policy

- Bot MUST NOT panic on any error
- All panics MUST be recovered and logged
- Bot MUST continue running after any non-fatal error

## 11. Security Requirements

### 11.1 User Validation

- All incoming updates MUST be validated for user information
- Only users in the whitelist may interact with the bot
- Unauthorized access attempts MUST be logged

### 11.2 Input Sanitization

- User messages are passed directly to LLM (LLM handles sanitization)
- No user input is logged to files (only error context)

### 11.3 Privacy

- User IDs are logged for security monitoring
- Message content is stored only in session files
- Session files MUST be stored in a directory that is gitignored

## 12. Configuration Integration

### 12.1 Required Config Structure

The bot reads from the following config fields:

```yaml
telegram:
  token: "BOT_TOKEN"  # From config.yaml

allowed_users:
  - 123456789        # List of allowed Telegram user IDs
  - 987654321

memory:
  path: "./data/sessions"
  max_messages: 50
```

### 12.2 Environment Variables

No additional environment variables are required for the bot layer beyond those needed for LLM providers.

## 13. Testing Requirements

### 13.1 Unit Tests

- Auth middleware MUST have tests for:
  - Whitelisted user allowed through
  - Non-whitelisted user blocked
  - Missing user information handled

- Command handlers MUST have tests for:
  - Each command returns correct response format
  - Commands are idempotent

- Message handler MUST have tests for:
  - Message flow with valid LLM response
  - Error handling for LLM failures

### 13.2 Mock Requirements

- LLM router MUST be mockable for testing
- Session manager MUST be mockable for testing

## 14. Dependencies

### 14.1 Required Packages

| Package | Version | Purpose |
|---------|---------|---------|
| github.com/go-telegram/bot | v1.18.0 | Telegram Bot API |

### 14.2 Internal Dependencies

| Module | Purpose |
|--------|---------|
| internal/config | Configuration loading |
| internal/llm | LLM routing |
| internal/session | Session management |

## 15. Performance Considerations

### 15.1 Timeouts

- LLM request timeout: 120 seconds (configurable)
- Bot message processing: Non-blocking per user

### 15.2 Concurrency

- Multiple users can interact with the bot concurrently
- Each user's session is isolated

## 16. Future Considerations

These features are out of scope for this specification but may be added later:
- Webhook support
- Group chat support
- Voice messages
- Image analysis
- Conversation branching