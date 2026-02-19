# Telegram Bot Implementation Checklist

Based on specification: `005_telegram_bot_spec.md`

## Pre-requisites

- [ ] Verify `internal/config` module is complete and functional
- [ ] Verify `internal/llm` module is complete and functional
- [ ] Verify `internal/session` module exists and is functional
- [ ] Verify `github.com/go-telegram/bot v1.18.0` is in go.mod

## Implementation Tasks

### Phase 1: Bot Entry Point

- [ ] Create `cmd/bot/main.go` with proper package imports
- [ ] Add config loading from config.yaml and .env
- [ ] Add validation for telegram token presence
- [ ] Add validation for at least one LLM provider enabled
- [ ] Initialize LLM router with config
- [ ] Initialize session manager with config.Memory
- [ ] Initialize Telegram bot with token
- [ ] Add startup logging for bot, allowed users count, and polling start

### Phase 2: Authentication Middleware

- [ ] Create `internal/bot/auth.go` file
- [ ] Implement AuthMiddleware function that wraps handler
- [ ] Add user ID extraction from Message updates
- [ ] Add user ID extraction from CallbackQuery updates
- [ ] Add user ID extraction from EditedMessage updates
- [ ] Add whitelist check against config.AllowedUsers
- [ ] Send "Access denied" message for unauthorized users
- [ ] Add logging for unauthorized access attempts
- [ ] Add edge case handling for missing user information
- [ ] Add development mode warning when allowed_users is empty

### Phase 3: Command Handlers

- [ ] Create `internal/bot/handlers.go` file
- [ ] Implement StartHandler for /start command with welcome message
- [ ] Implement HelpHandler for /help command with detailed help
- [ ] Implement MyIDHandler for /myid command returning user's Telegram ID
- [ ] Implement ModelHandler for /model command showing active provider and all providers
- [ ] Implement ClearHandler for /clear command with session deletion
- [ ] Add idempotency handling for /clear command
- [ ] Add error handling for session deletion failures

### Phase 4: Message Handler

- [ ] Implement TextMessageHandler for non-command text messages
- [ ] Add session loading from session manager
- [ ] Add user message to conversation history
- [ ] Add call to LLM router SendMessage
- [ ] Add assistant response to conversation history
- [ ] Add session saving to session manager
- [ ] Add typing indicator while waiting for LLM response
- [ ] Add success response handling - send LLM response to user
- [ ] Add no provider enabled error handling
- [ ] Add LLM API error handling with user-friendly message
- [ ] Add timeout error handling
- [ ] Add empty response handling

### Phase 5: Error Handling and Logging

- [ ] Add logging for all errors with timestamp, user ID, context, and message
- [ ] Add user-friendly error messages for each error type
- [ ] Add panic recovery to prevent bot crashes
- [ ] Ensure bot continues running after non-fatal errors

### Phase 6: Testing

- [ ] Write unit tests for auth middleware (whitelist validation)
- [ ] Write unit tests for command handlers (response format verification)
- [ ] Write unit tests for message handler (success and error flows)
- [ ] Ensure LLM router is mockable for testing
- [ ] Ensure session manager is mockable for testing

## Verification Steps

- [ ] Bot starts without errors when config is valid
- [ ] Unauthorized users receive "Access denied" message
- [ ] /start command returns welcome message with command list
- [ ] /help command returns detailed help message
- [ ] /myid command returns user's Telegram ID
- [ ] /model command returns active provider and all providers
- [ ] /clear command clears session and returns confirmation
- [ ] Text messages are processed through LLM and response is returned
- [ ] LLM errors are handled gracefully with user-friendly messages
- [ ] Multiple users can interact concurrently without conflicts

## Implementation Order Rationale

1. Entry point first - establishes the bot lifecycle
2. Auth middleware second - security is foundational
3. Command handlers third - simpler, synchronous operations
4. Message handler fourth - depends on LLM and session integration
5. Error handling - ensures robustness across all components
6. Testing last - validates all implemented components