package llm

import (
	"fmt"

	"github.com/jrswab/helpi/internal/config"
)

func NewProvider(cfg *config.Config, providerType string) (Provider, error) {
	switch providerType {
	case "openai":
		return NewOpenAIProvider(cfg), nil
	case "anthropic":
		return NewAnthropicProvider(cfg), nil
	case "ollama":
		return NewOllamaProvider(cfg), nil
	case "openrouter":
		return NewOpenRouterProvider(cfg), nil
	case "opencode":
		return NewOpenCodeProvider(cfg), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}

func NewRouter(cfg *config.Config) (Router, error) {
	providers := []Provider{}
	defaultIdx := -1

	if cfg.Providers.OpenAI.Enabled {
		providers = append(providers, NewOpenAIProvider(cfg))
		if defaultIdx == -1 {
			defaultIdx = len(providers) - 1
		}
	}

	if cfg.Providers.Anthropic.Enabled {
		providers = append(providers, NewAnthropicProvider(cfg))
		if defaultIdx == -1 {
			defaultIdx = len(providers) - 1
		}
	}

	if cfg.Providers.Ollama.Enabled {
		providers = append(providers, NewOllamaProvider(cfg))
		if defaultIdx == -1 {
			defaultIdx = len(providers) - 1
		}
	}

	if cfg.Providers.OpenRouter.Enabled {
		providers = append(providers, NewOpenRouterProvider(cfg))
		if defaultIdx == -1 {
			defaultIdx = len(providers) - 1
		}
	}

	if cfg.Providers.OpenCode.Enabled {
		providers = append(providers, NewOpenCodeProvider(cfg))
		if defaultIdx == -1 {
			defaultIdx = len(providers) - 1
		}
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no LLM provider enabled")
	}

	return newRouter(providers, defaultIdx), nil
}
