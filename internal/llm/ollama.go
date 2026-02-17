package llm

import (
	"context"
	"fmt"

	"github.com/jrswab/helpi/internal/config"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

type ollamaProvider struct {
	client  openai.Client
	model   string
	baseURL string
	enabled bool
}

func NewOllamaProvider(cfg *config.Config) Provider {
	enabled := cfg.Providers.Ollama.Enabled

	baseURL := cfg.Providers.Ollama.DefaultModel
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1"
	}

	var client openai.Client
	if enabled {
		client = openai.NewClient(
			option.WithBaseURL(baseURL),
			option.WithAPIKey("ollama"),
		)
	}

	return &ollamaProvider{
		client:  client,
		model:   cfg.Providers.Ollama.DefaultModel,
		baseURL: baseURL,
		enabled: enabled,
	}
}

func (p *ollamaProvider) Name() string {
	return "ollama"
}

func (p *ollamaProvider) IsEnabled() bool {
	return p.enabled
}

func (p *ollamaProvider) SendMessage(ctx context.Context, messages []Message) (string, error) {
	if !p.enabled {
		return "", fmt.Errorf("ollama: provider not enabled")
	}

	openAIMessages := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	for i, msg := range messages {
		switch msg.Role {
		case "system":
			openAIMessages[i] = openai.SystemMessage(msg.Content)
		case "user":
			openAIMessages[i] = openai.UserMessage(msg.Content)
		case "assistant":
			openAIMessages[i] = openai.AssistantMessage(msg.Content)
		default:
			openAIMessages[i] = openai.UserMessage(msg.Content)
		}
	}

	resp, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(p.model),
		Messages: openAIMessages,
	})
	if err != nil {
		return "", fmt.Errorf("ollama: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", nil
	}

	return resp.Choices[0].Message.Content, nil
}
