package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/jrswab/helpi/internal/config"
)

type anthropicProvider struct {
	client      anthropic.Client
	model       string
	enabled     bool
	providerCfg config.ProviderConfig
}

func NewAnthropicProvider(cfg *config.Config) Provider {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	enabled := cfg.Providers.Anthropic.Enabled && apiKey != ""

	var client anthropic.Client
	if enabled {
		client = anthropic.NewClient(
			option.WithAPIKey(apiKey),
		)
	}

	return &anthropicProvider{
		client:      client,
		model:       cfg.Providers.Anthropic.DefaultModel,
		enabled:     enabled,
		providerCfg: cfg.Providers.Anthropic,
	}
}

func (p *anthropicProvider) Name() string {
	return "anthropic"
}

func (p *anthropicProvider) IsEnabled() bool {
	return p.enabled
}

func (p *anthropicProvider) SendMessage(ctx context.Context, messages []Message) (string, error) {
	if !p.enabled {
		return "", fmt.Errorf("anthropic: provider not enabled")
	}

	var systemMsg string
	var conversationMessages []anthropic.MessageParam

	for _, msg := range messages {
		if msg.Role == "system" {
			systemMsg = msg.Content
			continue
		}

		var role anthropic.MessageParamRole
		if msg.Role == "assistant" {
			role = anthropic.MessageParamRole("assistant")
		} else {
			role = anthropic.MessageParamRoleUser
		}

		msgParam := anthropic.MessageParam{
			Role: role,
			Content: []anthropic.ContentBlockParamUnion{
				{OfText: &anthropic.TextBlockParam{Text: msg.Content}},
			},
		}
		conversationMessages = append(conversationMessages, msgParam)
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: 1024,
	}

	if systemMsg != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: systemMsg},
		}
	}

	if len(conversationMessages) > 0 {
		params.Messages = conversationMessages
	}

	message, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("anthropic: %w", err)
	}

	if len(message.Content) == 0 {
		return "", nil
	}

	var responseText string
	for _, content := range message.Content {
		textBlock := content.AsText()
		responseText += textBlock.Text
	}

	return responseText, nil
}
