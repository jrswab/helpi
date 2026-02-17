package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENCODE_API_KEY")
	os.Unsetenv("OLLAMA_BASE_URL")

	dir := t.TempDir()

	configContent := `telegram:
  token: "test-token"
allowed_users:
  - 123456789
  - 987654321
providers:
  openai:
    enabled: true
    default_model: "gpt-4"
  anthropic:
    enabled: false
    default_model: ""
  openrouter:
    enabled: false
  opencode:
    enabled: false
  ollama:
    enabled: false
memory:
  path: "./data/sessions"
  max_messages: 50
`
	envContent := `TELEGRAM_BOT_TOKEN=test-token-from-env
OPENAI_API_KEY=test-api-key
`

	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Telegram.Token != "test-token-from-env" {
		t.Errorf("expected token from env, got %s", cfg.Telegram.Token)
	}

	if len(cfg.AllowedUsers) != 2 {
		t.Errorf("expected 2 allowed users, got %d", len(cfg.AllowedUsers))
	}

	if !cfg.Providers.OpenAI.Enabled {
		t.Error("expected openai to be enabled")
	}

	if cfg.Providers.OpenAI.DefaultModel != "gpt-4" {
		t.Errorf("expected openai default model to be gpt-4, got %s", cfg.Providers.OpenAI.DefaultModel)
	}

	if cfg.Memory.Path != "./data/sessions" {
		t.Errorf("expected memory path to be ./data/sessions, got %s", cfg.Memory.Path)
	}

	if cfg.Memory.MaxMessages != 50 {
		t.Errorf("expected max_messages to be 50, got %d", cfg.Memory.MaxMessages)
	}
}

func TestLoad_MissingConfigFile(t *testing.T) {
	dir := t.TempDir()
	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	_, err := Load()
	if err == nil {
		t.Error("expected error when config.yaml is missing")
	}
}

func TestLoad_MissingTelegramToken(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENCODE_API_KEY")
	os.Unsetenv("OLLAMA_BASE_URL")

	dir := t.TempDir()

	configContent := `telegram:
  token: ""
allowed_users:
  - 123456789
providers:
  openai:
    enabled: false
  anthropic:
    enabled: false
  openrouter:
    enabled: false
  opencode:
    enabled: false
  ollama:
    enabled: false
memory:
  path: "./data/sessions"
  max_messages: 50
`
	envContent := `TELEGRAM_BOT_TOKEN=
`

	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	_, err := Load()
	if err == nil {
		t.Error("expected error when telegram token is empty")
	}
	if err != nil && !strings.Contains(err.Error(), "telegram.token") {
		t.Errorf("expected error to mention telegram.token, got: %v", err)
	}
}

func TestLoad_EmptyAllowedUsers(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENCODE_API_KEY")
	os.Unsetenv("OLLAMA_BASE_URL")

	dir := t.TempDir()

	configContent := `telegram:
  token: "test-token"
allowed_users: []
providers:
  openai:
    enabled: false
  anthropic:
    enabled: false
  openrouter:
    enabled: false
  opencode:
    enabled: false
  ollama:
    enabled: false
memory:
  path: "./data/sessions"
  max_messages: 50
`
	envContent := `TELEGRAM_BOT_TOKEN=test-token
`

	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.AllowedUsers != nil && len(cfg.AllowedUsers) != 0 {
		t.Error("expected empty allowed_users array to be valid")
	}
}

func TestLoad_AllProvidersDisabled(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENCODE_API_KEY")
	os.Unsetenv("OLLAMA_BASE_URL")

	dir := t.TempDir()

	configContent := `telegram:
  token: "test-token"
allowed_users:
  - 123456789
providers:
  openai:
    enabled: false
  anthropic:
    enabled: false
  openrouter:
    enabled: false
  opencode:
    enabled: false
  ollama:
    enabled: false
memory:
  path: "./data/sessions"
  max_messages: 50
`
	envContent := `TELEGRAM_BOT_TOKEN=test-token
`

	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	_, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
}

func TestLoad_MissingAPIKeyForEnabledProvider(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENCODE_API_KEY")
	os.Unsetenv("OLLAMA_BASE_URL")

	dir := t.TempDir()

	configContent := `telegram:
  token: "test-token"
allowed_users:
  - 123456789
providers:
  openai:
    enabled: true
    default_model: "gpt-4"
  anthropic:
    enabled: false
  openrouter:
    enabled: false
  opencode:
    enabled: false
  ollama:
    enabled: false
memory:
  path: "./data/sessions"
  max_messages: 50
`
	envContent := `TELEGRAM_BOT_TOKEN=test-token
`

	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	_, err := Load()
	if err == nil {
		t.Error("expected error when enabled provider has missing API key")
	}
	if !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Errorf("expected error to mention OPENAI_API_KEY, got: %v", err)
	}
}

func TestLoad_UnknownFieldsIgnored(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENCODE_API_KEY")
	os.Unsetenv("OLLAMA_BASE_URL")

	dir := t.TempDir()

	configContent := `telegram:
  token: "test-token"
allowed_users:
  - 123456789
providers:
  openai:
    enabled: false
  anthropic:
    enabled: false
  openrouter:
    enabled: false
  opencode:
    enabled: false
  ollama:
    enabled: false
memory:
  path: "./data/sessions"
  max_messages: 50
unknown_field: "should be ignored"
another_unknown: 12345
`
	envContent := `TELEGRAM_BOT_TOKEN=test-token
`

	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	_, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
}

func TestLoad_WhitespaceTrimming(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENCODE_API_KEY")
	os.Unsetenv("OLLAMA_BASE_URL")

	dir := t.TempDir()

	configContent := `telegram:
  token: "  test-token-with-spaces  "
allowed_users:
  - 123456789
providers:
  openai:
    enabled: false
  anthropic:
    enabled: false
  openrouter:
    enabled: false
  opencode:
    enabled: false
  ollama:
    enabled: false
memory:
  path: "  ./data/sessions  "
  max_messages: 50
`

	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}

	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Telegram.Token != "test-token-with-spaces" {
		t.Errorf("expected token to be trimmed, got %s", cfg.Telegram.Token)
	}
}

func TestLoad_DefaultMemoryValues(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENCODE_API_KEY")
	os.Unsetenv("OLLAMA_BASE_URL")

	dir := t.TempDir()

	configContent := `telegram:
  token: "test-token"
allowed_users:
  - 123456789
providers:
  openai:
    enabled: false
  anthropic:
    enabled: false
  openrouter:
    enabled: false
  opencode:
    enabled: false
  ollama:
    enabled: false
memory:
  path: "./custom/path"
  max_messages: 100
`
	envContent := `TELEGRAM_BOT_TOKEN=test-token
`

	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Memory.Path != "./custom/path" {
		t.Errorf("expected custom memory path, got %s", cfg.Memory.Path)
	}

	if cfg.Memory.MaxMessages != 100 {
		t.Errorf("expected custom max_messages, got %d", cfg.Memory.MaxMessages)
	}
}

func TestLoad_MaxMessagesValidation(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENCODE_API_KEY")
	os.Unsetenv("OLLAMA_BASE_URL")

	dir := t.TempDir()

	configContent := `telegram:
  token: "test-token"
allowed_users:
  - 123456789
providers:
  openai:
    enabled: false
  anthropic:
    enabled: false
  openrouter:
    enabled: false
  opencode:
    enabled: false
  ollama:
    enabled: false
memory:
  path: "./data/sessions"
  max_messages: 0
`
	envContent := `TELEGRAM_BOT_TOKEN=test-token
`

	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	_, err := Load()
	if err == nil {
		t.Error("expected error when max_messages is less than 1")
	}
}

func TestLoad_EnabledProviderWithoutModel(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENCODE_API_KEY")
	os.Unsetenv("OLLAMA_BASE_URL")

	dir := t.TempDir()

	configContent := `telegram:
  token: "test-token"
allowed_users:
  - 123456789
providers:
  openai:
    enabled: true
    default_model: ""
  anthropic:
    enabled: false
  openrouter:
    enabled: false
  opencode:
    enabled: false
  ollama:
    enabled: false
memory:
  path: "./data/sessions"
  max_messages: 50
`
	envContent := `TELEGRAM_BOT_TOKEN=test-token
OPENAI_API_KEY=test-key
`

	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	_, err := Load()
	if err == nil {
		t.Error("expected error when enabled provider has empty default_model")
	}
}

func TestLoad_OllamaDefaults(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENCODE_API_KEY")
	os.Unsetenv("OLLAMA_BASE_URL")

	dir := t.TempDir()

	configContent := `telegram:
  token: "test-token"
allowed_users:
  - 123456789
providers:
  openai:
    enabled: false
  anthropic:
    enabled: false
  openrouter:
    enabled: false
  opencode:
    enabled: false
  ollama:
    enabled: true
    default_model: "llama2"
memory:
  path: "./data/sessions"
  max_messages: 50
`
	envContent := `TELEGRAM_BOT_TOKEN=test-token
`

	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.APIKeys["OLLAMA_BASE_URL"] != "http://localhost:11434" {
		t.Errorf("expected default Ollama URL, got %s", cfg.APIKeys["OLLAMA_BASE_URL"])
	}
}
