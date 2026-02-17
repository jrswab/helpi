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

type openAIProvider struct {
	client      openai.Client
	model       string
	enabled     bool
	providerCfg config.ProviderConfig
}

func NewOpenAIProvider(cfg *config.Config) Provider {
	apiKey := os.Getenv("OPENAI_API_KEY")
	enabled := cfg.Providers.OpenAI.Enabled && apiKey != ""

	var client openai.Client
	if enabled {
		client = openai.NewClient(
			option.WithAPIKey(apiKey),
		)
	}

	return &openAIProvider{
		client:      client,
		model:       cfg.Providers.OpenAI.DefaultModel,
		enabled:     enabled,
		providerCfg: cfg.Providers.OpenAI,
	}
}

func (p *openAIProvider) Name() string {
	return "openai"
}

func (p *openAIProvider) IsEnabled() bool {
	return p.enabled
}

func (p *openAIProvider) SendMessage(ctx context.Context, messages []Message) (string, error) {
	if !p.enabled {
		return "", fmt.Errorf("openai: provider not enabled")
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
		return "", fmt.Errorf("openai: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", nil
	}

	return resp.Choices[0].Message.Content, nil
}
