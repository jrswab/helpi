package llm

import (
	"os"
	"testing"

	"github.com/jrswab/helpi/internal/config"
)

func TestOpenRouterProvider_Name(t *testing.T) {
	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			OpenRouter: config.ProviderConfig{
				Enabled:      false,
				DefaultModel: "openai/gpt-4o",
			},
		},
	}

	provider := NewOpenRouterProvider(cfg)

	if provider.Name() != "openrouter" {
		t.Errorf("Name() = %v, want openrouter", provider.Name())
	}
}

func TestOpenRouterProvider_IsEnabled_EnabledWithAPIKey(t *testing.T) {
	os.Setenv("OPENROUTER_API_KEY", "test-api-key")
	defer os.Unsetenv("OPENROUTER_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			OpenRouter: config.ProviderConfig{
				Enabled:      true,
				DefaultModel: "openai/gpt-4o",
			},
		},
	}

	provider := NewOpenRouterProvider(cfg)

	if !provider.IsEnabled() {
		t.Error("IsEnabled() = false, want true when enabled and API key present")
	}
}

func TestOpenRouterProvider_IsEnabled_Disabled(t *testing.T) {
	os.Unsetenv("OPENROUTER_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			OpenRouter: config.ProviderConfig{
				Enabled:      false,
				DefaultModel: "openai/gpt-4o",
			},
		},
	}

	provider := NewOpenRouterProvider(cfg)

	if provider.IsEnabled() {
		t.Error("IsEnabled() = true, want false when disabled")
	}
}
