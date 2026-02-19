package llm

import (
	"context"
	"os"
	"testing"

	"github.com/jrswab/helpi/internal/config"
)

func TestOpenCodeProvider_Name(t *testing.T) {
	os.Setenv("OPENCODE_API_KEY", "test-api-key")
	defer os.Unsetenv("OPENCODE_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			OpenCode: config.ProviderConfig{
				Enabled:      true,
				DefaultModel: "opencode-model",
			},
		},
	}

	provider := NewOpenCodeProvider(cfg)

	if provider.Name() != "opencode" {
		t.Errorf("Name() = %v, want opencode", provider.Name())
	}
}

func TestOpenCodeProvider_IsEnabled_EnabledWithAPIKey(t *testing.T) {
	os.Setenv("OPENCODE_API_KEY", "test-api-key")
	defer os.Unsetenv("OPENCODE_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			OpenCode: config.ProviderConfig{
				Enabled:      true,
				DefaultModel: "opencode-model",
			},
		},
	}

	provider := NewOpenCodeProvider(cfg)

	if !provider.IsEnabled() {
		t.Error("IsEnabled() = false, want true when enabled and API key present")
	}
}

func TestOpenCodeProvider_IsEnabled_Disabled(t *testing.T) {
	os.Unsetenv("OPENCODE_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			OpenCode: config.ProviderConfig{
				Enabled:      false,
				DefaultModel: "opencode-model",
			},
		},
	}

	provider := NewOpenCodeProvider(cfg)

	if provider.IsEnabled() {
		t.Error("IsEnabled() = true, want false when disabled")
	}
}

func TestOpenCodeProvider_SendMessage_Disabled(t *testing.T) {
	os.Unsetenv("OPENCODE_API_KEY")

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			OpenCode: config.ProviderConfig{
				Enabled:      false,
				DefaultModel: "opencode-model",
			},
		},
	}

	provider := NewOpenCodeProvider(cfg)

	_, err := provider.SendMessage(context.Background(), []Message{
		{Role: "user", Content: "Hello"},
	})

	if err == nil {
		t.Error("SendMessage() error = nil, want error when provider disabled")
	}

	expectedErr := "opencode: provider not enabled"
	if err.Error() != expectedErr {
		t.Errorf("SendMessage() error = %v, want %v", err.Error(), expectedErr)
	}
}
