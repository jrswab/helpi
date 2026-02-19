package llm

import (
	"os"
	"strings"
	"testing"

	"github.com/jrswab/helpi/internal/config"
)

func TestNewProvider(t *testing.T) {
	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			OpenAI:     config.ProviderConfig{Enabled: true, DefaultModel: "gpt-4"},
			Anthropic:  config.ProviderConfig{Enabled: true, DefaultModel: "claude-3"},
			Ollama:     config.ProviderConfig{Enabled: true, DefaultModel: "llama2"},
			OpenRouter: config.ProviderConfig{Enabled: true, DefaultModel: "openrouter-model"},
			OpenCode:   config.ProviderConfig{Enabled: true, DefaultModel: "opencode-model"},
		},
		APIKeys: map[string]string{
			"OPENAI_API_KEY":     "test-key",
			"ANTHROPIC_API_KEY":  "test-key",
			"OPENROUTER_API_KEY": "test-key",
			"OPENCODE_API_KEY":   "test-key",
			"OLLAMA_BASE_URL":    "http://localhost:11434",
		},
	}

	tests := []struct {
		name         string
		providerType string
		wantProvider string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "openai returns OpenAI provider",
			providerType: "openai",
			wantProvider: "openai",
		},
		{
			name:         "anthropic returns Anthropic provider",
			providerType: "anthropic",
			wantProvider: "anthropic",
		},
		{
			name:         "ollama returns Ollama provider",
			providerType: "ollama",
			wantProvider: "ollama",
		},
		{
			name:         "openrouter returns OpenRouter provider",
			providerType: "openrouter",
			wantProvider: "openrouter",
		},
		{
			name:         "opencode returns OpenCode provider",
			providerType: "opencode",
			wantProvider: "opencode",
		},
		{
			name:         "unknown returns error",
			providerType: "unknown",
			wantErr:      true,
			errContains:  "unknown provider type: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(cfg, tt.providerType)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if provider.Name() != tt.wantProvider {
				t.Errorf("expected provider name %q, got %q", tt.wantProvider, provider.Name())
			}
		})
	}
}

func TestNewRouter(t *testing.T) {
	unsetAllEnv := func() {
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
		os.Unsetenv("OPENROUTER_API_KEY")
		os.Unsetenv("OPENCODE_API_KEY")
		os.Unsetenv("OLLAMA_BASE_URL")
	}

	t.Run("multiple providers returns router", func(t *testing.T) {
		unsetAllEnv()
		os.Setenv("OPENAI_API_KEY", "test-key")
		os.Setenv("ANTHROPIC_API_KEY", "test-key")
		defer unsetAllEnv()

		cfg := &config.Config{
			Providers: config.ProvidersConfig{
				OpenAI:    config.ProviderConfig{Enabled: true, DefaultModel: "gpt-4"},
				Anthropic: config.ProviderConfig{Enabled: true, DefaultModel: "claude-3"},
				Ollama:    config.ProviderConfig{Enabled: false},
			},
			APIKeys: map[string]string{
				"OPENAI_API_KEY":    "test-key",
				"ANTHROPIC_API_KEY": "test-key",
			},
		}

		router, err := NewRouter(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if router == nil {
			t.Fatal("expected router, got nil")
		}
	})

	t.Run("first enabled provider becomes default", func(t *testing.T) {
		unsetAllEnv()
		os.Setenv("OPENAI_API_KEY", "test-key")
		os.Setenv("ANTHROPIC_API_KEY", "test-key")
		defer unsetAllEnv()

		cfg := &config.Config{
			Providers: config.ProvidersConfig{
				OpenAI:    config.ProviderConfig{Enabled: true, DefaultModel: "gpt-4"},
				Anthropic: config.ProviderConfig{Enabled: true, DefaultModel: "claude-3"},
			},
			APIKeys: map[string]string{
				"OPENAI_API_KEY":    "test-key",
				"ANTHROPIC_API_KEY": "test-key",
			},
		}

		router, err := NewRouter(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		defaultProvider, err := router.GetProvider()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if defaultProvider.Name() != "openai" {
			t.Errorf("expected default provider to be openai, got %s", defaultProvider.Name())
		}
	})

	t.Run("ollama only returns router", func(t *testing.T) {
		unsetAllEnv()
		os.Setenv("OLLAMA_BASE_URL", "http://localhost:11434")
		defer unsetAllEnv()

		cfg := &config.Config{
			Providers: config.ProvidersConfig{
				OpenAI:    config.ProviderConfig{Enabled: false},
				Anthropic: config.ProviderConfig{Enabled: false},
				Ollama:    config.ProviderConfig{Enabled: true, DefaultModel: "llama2"},
			},
			APIKeys: map[string]string{
				"OLLAMA_BASE_URL": "http://localhost:11434",
			},
		}

		router, err := NewRouter(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if router == nil {
			t.Fatal("expected router, got nil")
		}
	})

	t.Run("no providers returns error", func(t *testing.T) {
		unsetAllEnv()
		defer unsetAllEnv()

		cfg := &config.Config{
			Providers: config.ProvidersConfig{
				OpenAI:    config.ProviderConfig{Enabled: false},
				Anthropic: config.ProviderConfig{Enabled: false},
				Ollama:    config.ProviderConfig{Enabled: false},
			},
			APIKeys: map[string]string{},
		}

		router, err := NewRouter(cfg)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if router != nil {
			t.Error("expected nil router when no providers enabled")
		}

		if err.Error() != "no LLM provider enabled" {
			t.Errorf("expected error %q, got %q", "no LLM provider enabled", err.Error())
		}
	})
}
