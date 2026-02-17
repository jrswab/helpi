package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/jrswab/helpi/internal/config"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

type openRouterProvider struct {
	client      openai.Client
	model       string
	enabled     bool
	providerCfg config.ProviderConfig
}

func NewOpenRouterProvider(cfg *config.Config) Provider {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	enabled := cfg.Providers.OpenRouter.Enabled && apiKey != ""

	var client openai.Client
	if enabled {
		client = openai.NewClient(
			option.WithBaseURL("https://openrouter.ai/api/v1"),
			option.WithAPIKey(apiKey),
			option.WithHeader("HTTP-Referer", "https://github.com/jrswab/helpi"),
			option.WithHeader("X-Title", "Helpi"),
		)
	}

	return &openRouterProvider{
		client:      client,
		model:       cfg.Providers.OpenRouter.DefaultModel,
		enabled:     enabled,
		providerCfg: cfg.Providers.OpenRouter,
	}
}

func (p *openRouterProvider) Name() string {
	return "openrouter"
}

func (p *openRouterProvider) IsEnabled() bool {
	return p.enabled
}

func (p *openRouterProvider) SendMessage(ctx context.Context, messages []Message) (string, error) {
	if !p.enabled {
		return "", fmt.Errorf("openrouter: provider not enabled")
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
		return "", fmt.Errorf("openrouter: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", nil
	}

	return resp.Choices[0].Message.Content, nil
}
