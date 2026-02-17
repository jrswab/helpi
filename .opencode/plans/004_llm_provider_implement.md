# LLM Provider Implementation Plan

## Dependencies

- [x] Run `go get github.com/openai/openai-go/v3` to add OpenAI SDK
- [x] Run `go get github.com/anthropics/anthropic-sdk-go` to add Anthropic SDK
- [x] Run `go mod tidy` to verify dependencies

## Directory Structure

- [x] Create `internal/llm/` directory

## Core Types and Interfaces

- [x] Create `internal/llm/types.go` with `Message` struct
- [x] Create `internal/llm/provider.go` with `Provider` interface definition

## Provider Implementations

- [x] Implement `internal/llm/openai.go` with OpenAI provider
- [x] Implement `internal/llm/anthropic.go` with Anthropic provider
- [x] Implement `internal/llm/ollama.go` with Ollama provider
- [x] Implement `internal/llm/openrouter.go` with OpenRouter provider
- [x] Implement `internal/llm/opencode.go` with OpenCode Zen provider

## Router and Factory

- [x] Create `internal/llm/router.go` with Router interface and implementation
- [x] Create `internal/llm/factory.go` with `NewProvider()` and `NewRouter()` functions

## Integration

- [x] Add LLM package exports to `internal/llm/` package main file
- [x] Verify router integrates with existing config loader