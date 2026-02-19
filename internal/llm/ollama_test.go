package llm

import (
	"context"
	"os"
	"testing"

	"github.com/jrswab/helpi/internal/config"
)

func TestOllamaProvider_Name(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENCODE_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			Ollama: config.ProviderConfig{
				Enabled:      true,
				DefaultModel: "llama3.2",
			},
		},
	}

	provider := NewOllamaProvider(cfg)

	if provider.Name() != "ollama" {
		t.Errorf("Name() = %v, want ollama", provider.Name())
	}
}

func TestOllamaProvider_IsEnabled_Enabled(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENCODE_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			Ollama: config.ProviderConfig{
				Enabled:      true,
				DefaultModel: "llama3.2",
			},
		},
	}

	provider := NewOllamaProvider(cfg)

	if !provider.IsEnabled() {
		t.Error("IsEnabled() = false, want true when enabled")
	}
}

func TestOllamaProvider_IsEnabled_Disabled(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENCODE_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			Ollama: config.ProviderConfig{
				Enabled:      false,
				DefaultModel: "llama3.2",
			},
		},
	}

	provider := NewOllamaProvider(cfg)

	if provider.IsEnabled() {
		t.Error("IsEnabled() = true, want false when disabled")
	}
}

func TestOllamaProvider_SendMessage_Disabled(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENCODE_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			Ollama: config.ProviderConfig{
				Enabled:      false,
				DefaultModel: "llama3.2",
			},
		},
	}

	provider := NewOllamaProvider(cfg)

	_, err := provider.SendMessage(context.Background(), []Message{
		{Role: "user", Content: "Hello"},
	})

	if err == nil {
		t.Error("SendMessage() error = nil, want error when provider disabled")
	}

	expectedErr := "ollama: provider not enabled"
	if err.Error() != expectedErr {
		t.Errorf("SendMessage() error = %v, want %v", err.Error(), expectedErr)
	}
}
