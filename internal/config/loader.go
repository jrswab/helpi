package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type ConfigError struct {
	Field   string
	Message string
	Path    string
}

func (e *ConfigError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: %s (path: %s)", e.Field, e.Message, e.Path)
	}
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

func Load() (*Config, error) {
	dir, err := findConfigDir()
	if err != nil {
		return nil, err
	}

	cfg, err := loadYAML(dir)
	if err != nil {
		return nil, err
	}

	if err := loadEnv(dir, cfg); err != nil {
		return nil, err
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	if cfg.Memory.Path == "" {
		cfg.Memory.Path = "./data/sessions"
	}
	if cfg.Memory.MaxMessages == 0 {
		cfg.Memory.MaxMessages = 50
	}

	return cfg, nil
}

func findConfigDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", &ConfigError{Message: "failed to get current working directory", Path: ""}
	}

	configPath := filepath.Join(cwd, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return cwd, nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return "", &ConfigError{Message: "failed to get executable path", Path: ""}
	}

	exeDir := filepath.Dir(exePath)
	exeConfigPath := filepath.Join(exeDir, "config.yaml")
	if _, err := os.Stat(exeConfigPath); err == nil {
		return exeDir, nil
	}

	return "", &ConfigError{Message: "config.yaml not found", Path: cwd}
}

func loadYAML(dir string) (*Config, error) {
	path := filepath.Join(dir, "config.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &ConfigError{Message: "config.yaml not found", Path: path}
		}
		return nil, &ConfigError{Message: "failed to read config file", Path: path}
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, &ConfigError{Message: "config.yaml is empty", Path: path}
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, &ConfigError{Message: fmt.Sprintf("failed to parse YAML: %v", err), Path: path}
	}

	cfg.APIKeys = make(map[string]string)

	return &cfg, nil
}

func loadEnv(dir string, cfg *Config) error {
	envPath := filepath.Join(dir, ".env")

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		if os.Getenv("TELEGRAM_BOT_TOKEN") != "" {
			cfg.Telegram.Token = os.Getenv("TELEGRAM_BOT_TOKEN")
		}
		return nil
	}

	if err := godotenv.Load(envPath); err != nil {
		return &ConfigError{Message: fmt.Sprintf("failed to parse .env file: %v", err), Path: envPath}
	}

	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		cfg.Telegram.Token = token
	}

	cfg.APIKeys["OPENAI_API_KEY"] = os.Getenv("OPENAI_API_KEY")
	cfg.APIKeys["ANTHROPIC_API_KEY"] = os.Getenv("ANTHROPIC_API_KEY")
	cfg.APIKeys["OPENROUTER_API_KEY"] = os.Getenv("OPENROUTER_API_KEY")
	cfg.APIKeys["OPENCODE_API_KEY"] = os.Getenv("OPENCODE_API_KEY")
	cfg.APIKeys["OLLAMA_BASE_URL"] = os.Getenv("OLLAMA_BASE_URL")

	return nil
}

func validateConfig(cfg *Config) error {
	if cfg.Telegram.Token = strings.TrimSpace(cfg.Telegram.Token); cfg.Telegram.Token == "" {
		return &ConfigError{Field: "telegram.token", Message: "is required and cannot be empty"}
	}

	if cfg.AllowedUsers == nil {
		return &ConfigError{Field: "allowed_users", Message: "is required and cannot be nil"}
	}

	for _, userID := range cfg.AllowedUsers {
		if userID <= 0 {
			return &ConfigError{Field: "allowed_users", Message: "each user ID must be a positive integer"}
		}
	}

	if cfg.Providers.OpenAI.Enabled && cfg.Providers.OpenAI.DefaultModel == "" {
		return &ConfigError{Field: "providers.openai.default_model", Message: "is required when provider is enabled"}
	}
	if cfg.Providers.Anthropic.Enabled && cfg.Providers.Anthropic.DefaultModel == "" {
		return &ConfigError{Field: "providers.anthropic.default_model", Message: "is required when provider is enabled"}
	}
	if cfg.Providers.OpenRouter.Enabled && cfg.Providers.OpenRouter.DefaultModel == "" {
		return &ConfigError{Field: "providers.openrouter.default_model", Message: "is required when provider is enabled"}
	}
	if cfg.Providers.OpenCode.Enabled && cfg.Providers.OpenCode.DefaultModel == "" {
		return &ConfigError{Field: "providers.opencode.default_model", Message: "is required when provider is enabled"}
	}

	if cfg.Memory.MaxMessages < 1 {
		return &ConfigError{Field: "memory.max_messages", Message: "must be >= 1"}
	}

	if err := validateAPIKeys(cfg); err != nil {
		return err
	}

	return nil
}

func validateAPIKeys(cfg *Config) error {
	if cfg.Providers.OpenAI.Enabled {
		if cfg.APIKeys["OPENAI_API_KEY"] == "" {
			return &ConfigError{Field: "OPENAI_API_KEY", Message: "is required when openai provider is enabled"}
		}
	}

	if cfg.Providers.Anthropic.Enabled {
		if cfg.APIKeys["ANTHROPIC_API_KEY"] == "" {
			return &ConfigError{Field: "ANTHROPIC_API_KEY", Message: "is required when anthropic provider is enabled"}
		}
	}

	if cfg.Providers.OpenRouter.Enabled {
		if cfg.APIKeys["OPENROUTER_API_KEY"] == "" {
			return &ConfigError{Field: "OPENROUTER_API_KEY", Message: "is required when openrouter provider is enabled"}
		}
	}

	if cfg.Providers.OpenCode.Enabled {
		if cfg.APIKeys["OPENCODE_API_KEY"] == "" {
			return &ConfigError{Field: "OPENCODE_API_KEY", Message: "is required when opencode provider is enabled"}
		}
	}

	if cfg.APIKeys["OLLAMA_BASE_URL"] == "" {
		cfg.APIKeys["OLLAMA_BASE_URL"] = "http://localhost:11434"
	}

	return nil
}
