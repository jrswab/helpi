package llm

import (
	"context"
	"os"
	"testing"

	"github.com/jrswab/helpi/internal/config"
)

func TestAnthropicProvider_Name(t *testing.T) {
	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			Anthropic: config.ProviderConfig{
				Enabled:      false,
				DefaultModel: "claude-3-5-sonnet-20241022",
			},
		},
	}

	provider := NewAnthropicProvider(cfg)

	if provider.Name() != "anthropic" {
		t.Errorf("Name() = %v, want anthropic", provider.Name())
	}
}

func TestAnthropicProvider_IsEnabled_EnabledWithAPIKey(t *testing.T) {
	os.Setenv("ANTHROPIC_API_KEY", "test-api-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			Anthropic: config.ProviderConfig{
				Enabled:      true,
				DefaultModel: "claude-3-5-sonnet-20241022",
			},
		},
	}

	provider := NewAnthropicProvider(cfg)

	if !provider.IsEnabled() {
		t.Error("IsEnabled() = false, want true when enabled and API key present")
	}
}

func TestAnthropicProvider_IsEnabled_Disabled(t *testing.T) {
	os.Unsetenv("ANTHROPIC_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			Anthropic: config.ProviderConfig{
				Enabled:      false,
				DefaultModel: "claude-3-5-sonnet-20241022",
			},
		},
	}

	provider := NewAnthropicProvider(cfg)

	if provider.IsEnabled() {
		t.Error("IsEnabled() = true, want false when disabled")
	}
}

func TestAnthropicProvider_SendMessage_Disabled(t *testing.T) {
	os.Unsetenv("ANTHROPIC_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			Anthropic: config.ProviderConfig{
				Enabled:      false,
				DefaultModel: "claude-3-5-sonnet-20241022",
			},
		},
	}

	provider := NewAnthropicProvider(cfg)

	_, err := provider.SendMessage(context.Background(), []Message{
		{Role: "user", Content: "Hello"},
	})

	if err == nil {
		t.Error("SendMessage() error = nil, want error when provider disabled")
	}

	expectedErr := "anthropic: provider not enabled"
	if err.Error() != expectedErr {
		t.Errorf("SendMessage() error = %v, want %v", err.Error(), expectedErr)
	}
}
