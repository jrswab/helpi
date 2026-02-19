package llm

import (
	"context"
	"os"
	"testing"

	"github.com/jrswab/helpi/internal/config"
)

func TestOpenAIProvider_Name(t *testing.T) {
	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			OpenAI: config.ProviderConfig{
				Enabled:      false,
				DefaultModel: "gpt-4o",
			},
		},
	}

	provider := NewOpenAIProvider(cfg)

	if provider.Name() != "openai" {
		t.Errorf("Name() = %v, want openai", provider.Name())
	}
}

func TestOpenAIProvider_IsEnabled_EnabledWithAPIKey(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-api-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			OpenAI: config.ProviderConfig{
				Enabled:      true,
				DefaultModel: "gpt-4o",
			},
		},
	}

	provider := NewOpenAIProvider(cfg)

	if !provider.IsEnabled() {
		t.Error("IsEnabled() = false, want true when enabled and API key present")
	}
}

func TestOpenAIProvider_IsEnabled_Disabled(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			OpenAI: config.ProviderConfig{
				Enabled:      false,
				DefaultModel: "gpt-4o",
			},
		},
	}

	provider := NewOpenAIProvider(cfg)

	if provider.IsEnabled() {
		t.Error("IsEnabled() = true, want false when disabled")
	}
}

func TestOpenAIProvider_IsEnabled_EnabledNoAPIKey(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			OpenAI: config.ProviderConfig{
				Enabled:      true,
				DefaultModel: "gpt-4o",
			},
		},
	}

	provider := NewOpenAIProvider(cfg)

	if provider.IsEnabled() {
		t.Error("IsEnabled() = true, want false when enabled but no API key")
	}
}

func TestOpenAIProvider_SendMessage_Disabled(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			OpenAI: config.ProviderConfig{
				Enabled:      false,
				DefaultModel: "gpt-4o",
			},
		},
	}

	provider := NewOpenAIProvider(cfg)

	_, err := provider.SendMessage(context.Background(), []Message{
		{Role: "user", Content: "Hello"},
	})

	if err == nil {
		t.Error("SendMessage() error = nil, want error when provider disabled")
	}

	expectedErr := "openai: provider not enabled"
	if err.Error() != expectedErr {
		t.Errorf("SendMessage() error = %v, want %v", err.Error(), expectedErr)
	}
}
