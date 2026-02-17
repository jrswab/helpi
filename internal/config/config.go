package config

type Config struct {
	Telegram     TelegramConfig    `yaml:"telegram"`
	AllowedUsers []int64           `yaml:"allowed_users"`
	Providers    ProvidersConfig   `yaml:"providers"`
	Memory       MemoryConfig      `yaml:"memory"`
	APIKeys      map[string]string `yaml:"-"`
}

type TelegramConfig struct {
	Token string `yaml:"token"`
}

type ProviderConfig struct {
	Enabled      bool   `yaml:"enabled"`
	DefaultModel string `yaml:"default_model"`
}

type ProvidersConfig struct {
	OpenAI     ProviderConfig `yaml:"openai"`
	Anthropic  ProviderConfig `yaml:"anthropic"`
	OpenRouter ProviderConfig `yaml:"openrouter"`
	OpenCode   ProviderConfig `yaml:"opencode"`
	Ollama     ProviderConfig `yaml:"ollama"`
}

type MemoryConfig struct {
	Path        string `yaml:"path"`
	MaxMessages int    `yaml:"max_messages"`
}
