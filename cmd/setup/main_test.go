package main

import (
	"bufio"
	"bytes"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMaskString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "****"},
		{"abc", "****"},
		{"abcd", "****"},
		{"abcde", "****bcde"},
		{"abcdef", "****cdef"},
		{"1234567890", "****7890"},
		{"a", "****"},
		{"ab", "****"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := maskString(tt.input)
			if result != tt.expected {
				t.Errorf("maskString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsProviderEnabled(t *testing.T) {
	tests := []struct {
		name      string
		providers ProvidersConfig
		expected  bool
	}{
		{"openai", ProvidersConfig{OpenAI: ProviderConfig{Enabled: true}}, true},
		{"openai", ProvidersConfig{OpenAI: ProviderConfig{Enabled: false}}, false},
		{"anthropic", ProvidersConfig{Anthropic: ProviderConfig{Enabled: true}}, true},
		{"anthropic", ProvidersConfig{Anthropic: ProviderConfig{Enabled: false}}, false},
		{"openrouter", ProvidersConfig{OpenRouter: ProviderConfig{Enabled: true}}, true},
		{"openrouter", ProvidersConfig{OpenRouter: ProviderConfig{Enabled: false}}, false},
		{"opencode", ProvidersConfig{OpenCode: ProviderConfig{Enabled: true}}, true},
		{"opencode", ProvidersConfig{OpenCode: ProviderConfig{Enabled: false}}, false},
		{"ollama", ProvidersConfig{Ollama: ProviderConfig{Enabled: true}}, true},
		{"ollama", ProvidersConfig{Ollama: ProviderConfig{Enabled: false}}, false},
		{"unknown", ProvidersConfig{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isProviderEnabled(tt.providers, tt.name)
			if result != tt.expected {
				t.Errorf("isProviderEnabled(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestSetProviderEnabled(t *testing.T) {
	providers := ProvidersConfig{}

	setProviderEnabled(&providers, "openai", true)
	if !providers.OpenAI.Enabled {
		t.Error("setProviderEnabled did not set OpenAI.Enabled to true")
	}

	setProviderEnabled(&providers, "anthropic", true)
	if !providers.Anthropic.Enabled {
		t.Error("setProviderEnabled did not set Anthropic.Enabled to true")
	}

	setProviderEnabled(&providers, "openai", false)
	if providers.OpenAI.Enabled {
		t.Error("setProviderEnabled did not set OpenAI.Enabled to false")
	}
}

func TestGetProviderModel(t *testing.T) {
	providers := ProvidersConfig{
		OpenAI:     ProviderConfig{DefaultModel: "gpt-4"},
		Anthropic:  ProviderConfig{DefaultModel: "claude-3"},
		OpenRouter: ProviderConfig{},
		OpenCode:   ProviderConfig{},
		Ollama:     ProviderConfig{},
	}

	tests := []struct {
		name     string
		expected string
	}{
		{"openai", "gpt-4"},
		{"anthropic", "claude-3"},
		{"openrouter", ""},
		{"opencode", ""},
		{"ollama", ""},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getProviderModel(providers, tt.name)
			if result != tt.expected {
				t.Errorf("getProviderModel(%q) = %q, want %q", tt.name, result, tt.expected)
			}
		})
	}
}

func TestSetProviderModel(t *testing.T) {
	providers := ProvidersConfig{}

	setProviderModel(&providers, "openai", "gpt-4o")
	if providers.OpenAI.DefaultModel != "gpt-4o" {
		t.Errorf("setProviderModel: OpenAI.DefaultModel = %q, want %q", providers.OpenAI.DefaultModel, "gpt-4o")
	}

	setProviderModel(&providers, "anthropic", "claude-3-5-sonnet")
	if providers.Anthropic.DefaultModel != "claude-3-5-sonnet" {
		t.Errorf("setProviderModel: Anthropic.DefaultModel = %q, want %q", providers.Anthropic.DefaultModel, "claude-3-5-sonnet")
	}

	setProviderModel(&providers, "ollama", "llama3.1")
	if providers.Ollama.DefaultModel != "llama3.1" {
		t.Errorf("setProviderModel: Ollama.DefaultModel = %q, want %q", providers.Ollama.DefaultModel, "llama3.1")
	}
}

func TestSaveConfig_YAMLMarshaling(t *testing.T) {
	tmpDir := t.TempDir()
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(origCwd)

	os.Chdir(tmpDir)

	cfg := &ExistingConfig{
		Telegram:     "test-token-123",
		AllowedUsers: []int64{123456789, 987654321},
		Providers: ProvidersConfig{
			OpenAI:    ProviderConfig{Enabled: true, DefaultModel: "gpt-4o"},
			Anthropic: ProviderConfig{Enabled: true, DefaultModel: "claude-3-5-sonnet"},
		},
		Memory: MemoryConfig{
			Path:        "./data/sessions",
			MaxMessages: 50,
		},
		APIKeys: map[string]string{
			"OPENAI_API_KEY": "sk-test-key",
		},
	}

	err = saveConfig(cfg)
	if err != nil {
		t.Fatalf("saveConfig failed: %v", err)
	}

	configData, err := os.ReadFile("config.yaml")
	if err != nil {
		t.Fatalf("failed to read config.yaml: %v", err)
	}

	var parsed map[string]interface{}
	err = yaml.Unmarshal(configData, &parsed)
	if err != nil {
		t.Fatalf("failed to unmarshal config.yaml: %v", err)
	}

	telegram, ok := parsed["telegram"].(map[string]interface{})
	if !ok {
		t.Fatal("telegram section not found in config")
	}
	if telegram["token"] != "test-token-123" {
		t.Errorf("telegram.token = %v, want test-token-123", telegram["token"])
	}

	providers, ok := parsed["providers"].(map[string]interface{})
	if !ok {
		t.Fatal("providers section not found in config")
	}
	openai, ok := providers["openai"].(map[string]interface{})
	if !ok {
		t.Fatal("openai provider not found in config")
	}
	if openai["enabled"] != true {
		t.Errorf("providers.openai.enabled = %v, want true", openai["enabled"])
	}
	if openai["default_model"] != "gpt-4o" {
		t.Errorf("providers.openai.default_model = %v, want gpt-4o", openai["default_model"])
	}

	envData, err := os.ReadFile(".env")
	if err != nil {
		t.Fatalf("failed to read .env: %v", err)
	}

	envContent := string(envData)
	if !contains(envContent, "TELEGRAM_BOT_TOKEN=test-token-123") {
		t.Error(".env does not contain TELEGRAM_BOT_TOKEN")
	}
	if !contains(envContent, "OPENAI_API_KEY=sk-test-key") {
		t.Error(".env does not contain OPENAI_API_KEY")
	}
}

func TestSaveConfig_WriteError(t *testing.T) {
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(origCwd)

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	os.WriteFile("config.yaml", []byte{}, 0000)
	os.WriteFile(".env", []byte{}, 0000)

	cfg := &ExistingConfig{
		Telegram: "test-token",
		APIKeys:  map[string]string{},
	}

	err = saveConfig(cfg)
	if err == nil {
		t.Error("saveConfig should fail with unwritable files")
	}
}

func TestProviderEnvKeys(t *testing.T) {
	expected := map[string]string{
		"openai":     "OPENAI_API_KEY",
		"anthropic":  "ANTHROPIC_API_KEY",
		"openrouter": "OPENROUTER_API_KEY",
		"opencode":   "OPENCODE_API_KEY",
		"ollama":     "OLLAMA_BASE_URL",
	}

	for k, v := range expected {
		if providerEnvKeys[k] != v {
			t.Errorf("providerEnvKeys[%q] = %q, want %q", k, providerEnvKeys[k], v)
		}
	}
}

func TestProviderDefaults(t *testing.T) {
	expected := map[string]string{
		"openai":     "gpt-4o",
		"anthropic":  "claude-3-5-sonnet-20241022",
		"openrouter": "openai/gpt-4o",
		"opencode":   "opencode/big-pickle",
		"ollama":     "llama3.2",
	}

	for k, v := range expected {
		if providerDefaults[k] != v {
			t.Errorf("providerDefaults[%q] = %q, want %q", k, providerDefaults[k], v)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestConfigStruct_Fields(t *testing.T) {
	cfg := ExistingConfig{
		Telegram:     "test-token",
		AllowedUsers: []int64{123},
		Providers: ProvidersConfig{
			OpenAI: ProviderConfig{Enabled: true, DefaultModel: "gpt-4o"},
		},
		Memory: MemoryConfig{
			Path:        "/tmp/test",
			MaxMessages: 100,
		},
		APIKeys: map[string]string{"key": "value"},
	}

	if cfg.Telegram != "test-token" {
		t.Errorf("Telegram field not set correctly")
	}
	if len(cfg.AllowedUsers) != 1 || cfg.AllowedUsers[0] != 123 {
		t.Errorf("AllowedUsers field not set correctly")
	}
	if !cfg.Providers.OpenAI.Enabled {
		t.Errorf("Providers.OpenAI.Enabled not set correctly")
	}
	if cfg.Memory.MaxMessages != 100 {
		t.Errorf("Memory.MaxMessages not set correctly")
	}
	if cfg.APIKeys["key"] != "value" {
		t.Errorf("APIKeys not set correctly")
	}
}

func TestReadLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "hello\n", "hello"},
		{"with spaces", "  hello world  \n", "hello world"},
		{"empty", "\n", ""},
		{"trailing newline removed", "test\n", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(bytes.NewReader([]byte(tt.input)))
			result := readLine(reader)
			if result != tt.expected {
				t.Errorf("readLine(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
