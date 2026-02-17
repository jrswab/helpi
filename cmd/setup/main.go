package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type ExistingConfig struct {
	Telegram     string            `yaml:"token" json:"token"`
	AllowedUsers []int64           `yaml:"allowed_users" json:"allowed_users"`
	Providers    ProvidersConfig   `yaml:"providers" json:"providers"`
	Memory       MemoryConfig      `yaml:"memory" json:"memory"`
	APIKeys      map[string]string `yaml:"-" json:"-"`
}

type ProvidersConfig struct {
	OpenAI     ProviderConfig `yaml:"openai" json:"openai"`
	Anthropic  ProviderConfig `yaml:"anthropic" json:"anthropic"`
	OpenRouter ProviderConfig `yaml:"openrouter" json:"openrouter"`
	OpenCode   ProviderConfig `yaml:"opencode" json:"opencode"`
	Ollama     ProviderConfig `yaml:"ollama" json:"ollama"`
}

type ProviderConfig struct {
	Enabled      bool   `yaml:"enabled" json:"enabled"`
	DefaultModel string `yaml:"default_model" json:"default_model"`
}

type MemoryConfig struct {
	Path        string `yaml:"path" json:"path"`
	MaxMessages int    `yaml:"max_messages" json:"max_messages"`
}

var providerDefaults = map[string]string{
	"openai":     "gpt-4o",
	"anthropic":  "claude-3-5-sonnet-20241022",
	"openrouter": "openai/gpt-4o",
	"opencode":   "opencode/big-pickle",
	"ollama":     "llama3.2",
}

var providerEnvKeys = map[string]string{
	"openai":     "OPENAI_API_KEY",
	"anthropic":  "ANTHROPIC_API_KEY",
	"openrouter": "OPENROUTER_API_KEY",
	"opencode":   "OPENCODE_API_KEY",
	"ollama":     "OLLAMA_BASE_URL",
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	cfg := &ExistingConfig{
		APIKeys: make(map[string]string),
		Providers: ProvidersConfig{
			OpenAI:     ProviderConfig{Enabled: false},
			Anthropic:  ProviderConfig{Enabled: false},
			OpenRouter: ProviderConfig{Enabled: false},
			OpenCode:   ProviderConfig{Enabled: false},
			Ollama:     ProviderConfig{Enabled: false},
		},
		Memory: MemoryConfig{
			Path:        "./data/sessions",
			MaxMessages: 50,
		},
	}

	loadExistingConfig(cfg)

	fmt.Println("=== Helpi Setup Wizard ===")
	fmt.Println()

	cfg.Telegram = promptToken(reader, cfg.Telegram)
	cfg.Providers = promptProviders(reader, cfg.Providers, cfg.APIKeys)
	cfg.AllowedUsers = promptAllowedUsers(reader, cfg.AllowedUsers)
	cfg.Memory = promptMemory(reader, cfg.Memory)

	if err := saveConfig(cfg); err != nil {
		fmt.Printf("✗ Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Configuration saved to config.yaml")
	fmt.Println("✓ Secrets saved to .env")
	fmt.Println()
	fmt.Println("Run the bot with: go run ./cmd/bot")
}

func readLine(reader *bufio.Reader) string {
	line, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(line)
}

func loadExistingConfig(cfg *ExistingConfig) {
	cwd, _ := os.Getwd()

	configPath := filepath.Join(cwd, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err == nil && len(strings.TrimSpace(string(data))) > 0 {
		yaml.Unmarshal(data, cfg)
	}

	envPath := filepath.Join(cwd, ".env")
	godotenv.Load(envPath)

	cfg.APIKeys["TELEGRAM_BOT_TOKEN"] = os.Getenv("TELEGRAM_BOT_TOKEN")
	cfg.APIKeys["OPENAI_API_KEY"] = os.Getenv("OPENAI_API_KEY")
	cfg.APIKeys["ANTHROPIC_API_KEY"] = os.Getenv("ANTHROPIC_API_KEY")
	cfg.APIKeys["OPENROUTER_API_KEY"] = os.Getenv("OPENROUTER_API_KEY")
	cfg.APIKeys["OPENCODE_API_KEY"] = os.Getenv("OPENCODE_API_KEY")
	cfg.APIKeys["OLLAMA_BASE_URL"] = os.Getenv("OLLAMA_BASE_URL")
}

func promptToken(reader *bufio.Reader, current string) string {
	for {
		prompt := "Telegram Bot Token (required): "
		if current != "" {
			prompt = fmt.Sprintf("Telegram Bot Token (required) [%s]: ", maskString(current))
		}
		fmt.Print(prompt)
		input := readLine(reader)

		if input != "" {
			return input
		}
		if current != "" {
			return current
		}
		fmt.Println("Error: Token is required")
	}
}

func promptProviders(reader *bufio.Reader, providers ProvidersConfig, apiKeys map[string]string) ProvidersConfig {
	providerList := []string{"openai", "anthropic", "openrouter", "opencode", "ollama"}

	for _, name := range providerList {
		enabled := isProviderEnabled(providers, name)
		defaultEnabled := "n"
		if enabled {
			defaultEnabled = "y"
		}

		for {
			fmt.Printf("Enable %s? (y/n) [%s]: ", strings.ToUpper(name[:1])+name[1:], defaultEnabled)
			input := readLine(reader)
			input = strings.ToLower(input)

			if input == "" {
				break
			}
			if input == "y" || input == "yes" {
				setProviderEnabled(&providers, name, true)
				break
			}
			if input == "n" || input == "no" {
				setProviderEnabled(&providers, name, false)
				break
			}
			fmt.Println("Please enter y or n")
		}

		if isProviderEnabled(providers, name) {
			envKey := providerEnvKeys[name]
			currentKey := apiKeys[envKey]

			apiKeys[envKey] = promptAPIKey(reader, name, currentKey, envKey)

			if apiKeys[envKey] == "" && envKey != "OLLAMA_BASE_URL" {
				fmt.Printf("Warning: %s is enabled but no API key is set - provider may not work\n", name)
			}

			defaultModel := providerDefaults[name]
			if currentModel := getProviderModel(providers, name); currentModel != "" {
				defaultModel = currentModel
			}

			setProviderModel(&providers, name, promptModel(reader, name, defaultModel))
		}
	}

	return providers
}

func promptAPIKey(reader *bufio.Reader, provider, current, envKey string) string {
	for {
		prompt := fmt.Sprintf("%s API Key: ", strings.ToUpper(provider[:1])+provider[1:])
		if current != "" {
			prompt = prompt + "[SET]: "
		} else if envKey == "OLLAMA_BASE_URL" {
			prompt = prompt + "[http://localhost:11434]: "
		}

		fmt.Print(prompt)
		input := readLine(reader)

		if input != "" && input != "[SET]" {
			return input
		}
		if current != "" {
			return current
		}
		if envKey == "OLLAMA_BASE_URL" {
			return "http://localhost:11434"
		}
		fmt.Printf("Error: %s API Key is required\n", provider)
	}
}

func promptModel(reader *bufio.Reader, provider, defaultModel string) string {
	prompt := fmt.Sprintf("%s default model [%s]: ", strings.ToUpper(provider[:1])+provider[1:], defaultModel)
	fmt.Print(prompt)
	input := readLine(reader)

	if input == "" {
		return defaultModel
	}
	return input
}

func promptAllowedUsers(reader *bufio.Reader, current []int64) []int64 {
	for {
		display := ""
		if len(current) > 0 {
			strs := make([]string, len(current))
			for i, id := range current {
				strs[i] = strconv.FormatInt(id, 10)
			}
			display = strings.Join(strs, ", ")
		}

		prompt := "Allowed Telegram User IDs (comma-separated): "
		if display != "" {
			prompt = prompt + "[" + display + "]: "
		} else {
			prompt = prompt + ": "
		}

		fmt.Print(prompt)
		input := readLine(reader)

		if input == "" {
			return current
		}

		parts := strings.Split(input, ",")
		var ids []int64
		seen := make(map[int64]bool)
		hasError := false

		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			id, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				fmt.Printf("Error: Invalid user ID: %s. Must be numeric.\n", part)
				hasError = true
				break
			}

			if !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}

		if hasError {
			continue
		}

		if len(ids) == 0 {
			return current
		}

		return ids
	}
}

func promptMemory(reader *bufio.Reader, memory MemoryConfig) MemoryConfig {
	prompt := fmt.Sprintf("Session storage path [default: %s]: ", memory.Path)
	fmt.Print(prompt)
	input := readLine(reader)
	if input != "" {
		memory.Path = input
	}

	fmt.Print(fmt.Sprintf("Max messages per conversation (0 to retain all) [default: %d]: ", memory.MaxMessages))
	input = readLine(reader)
	if input != "" {
		val, err := strconv.Atoi(input)
		if err != nil || val < 0 {
			fmt.Println("Error: Must be a non-negative integer")
			return promptMemory(reader, memory)
		}
		memory.MaxMessages = val
	}

	return memory
}

func saveConfig(cfg *ExistingConfig) error {
	yamlData := map[string]interface{}{
		"telegram": map[string]string{
			"token": cfg.Telegram,
		},
		"allowed_users": cfg.AllowedUsers,
		"providers":     cfg.Providers,
		"memory":        cfg.Memory,
	}

	data, err := yaml.Marshal(yamlData)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile("config.yaml", data, 0644); err != nil {
		return fmt.Errorf("failed to write config.yaml: %v", err)
	}

	envContent := ""
	if cfg.Telegram != "" {
		envContent += fmt.Sprintf("TELEGRAM_BOT_TOKEN=%s\n", cfg.Telegram)
	}
	if cfg.APIKeys["OPENAI_API_KEY"] != "" {
		envContent += fmt.Sprintf("OPENAI_API_KEY=%s\n", cfg.APIKeys["OPENAI_API_KEY"])
	}
	if cfg.APIKeys["ANTHROPIC_API_KEY"] != "" {
		envContent += fmt.Sprintf("ANTHROPIC_API_KEY=%s\n", cfg.APIKeys["ANTHROPIC_API_KEY"])
	}
	if cfg.APIKeys["OPENROUTER_API_KEY"] != "" {
		envContent += fmt.Sprintf("OPENROUTER_API_KEY=%s\n", cfg.APIKeys["OPENROUTER_API_KEY"])
	}
	if cfg.APIKeys["OPENCODE_API_KEY"] != "" {
		envContent += fmt.Sprintf("OPENCODE_API_KEY=%s\n", cfg.APIKeys["OPENCODE_API_KEY"])
	}
	if cfg.APIKeys["OLLAMA_BASE_URL"] != "" {
		envContent += fmt.Sprintf("OLLAMA_BASE_URL=%s\n", cfg.APIKeys["OLLAMA_BASE_URL"])
	}

	if err := os.WriteFile(".env", []byte(envContent), 0644); err != nil {
		return fmt.Errorf("failed to write .env: %v", err)
	}

	return nil
}

func isProviderEnabled(providers ProvidersConfig, name string) bool {
	switch name {
	case "openai":
		return providers.OpenAI.Enabled
	case "anthropic":
		return providers.Anthropic.Enabled
	case "openrouter":
		return providers.OpenRouter.Enabled
	case "opencode":
		return providers.OpenCode.Enabled
	case "ollama":
		return providers.Ollama.Enabled
	}
	return false
}

func setProviderEnabled(providers *ProvidersConfig, name string, enabled bool) {
	switch name {
	case "openai":
		providers.OpenAI.Enabled = enabled
	case "anthropic":
		providers.Anthropic.Enabled = enabled
	case "openrouter":
		providers.OpenRouter.Enabled = enabled
	case "opencode":
		providers.OpenCode.Enabled = enabled
	case "ollama":
		providers.Ollama.Enabled = enabled
	}
}

func getProviderModel(providers ProvidersConfig, name string) string {
	switch name {
	case "openai":
		return providers.OpenAI.DefaultModel
	case "anthropic":
		return providers.Anthropic.DefaultModel
	case "openrouter":
		return providers.OpenRouter.DefaultModel
	case "opencode":
		return providers.OpenCode.DefaultModel
	case "ollama":
		return providers.Ollama.DefaultModel
	}
	return ""
}

func setProviderModel(providers *ProvidersConfig, name, model string) {
	switch name {
	case "openai":
		providers.OpenAI.DefaultModel = model
	case "anthropic":
		providers.Anthropic.DefaultModel = model
	case "openrouter":
		providers.OpenRouter.DefaultModel = model
	case "opencode":
		providers.OpenCode.DefaultModel = model
	case "ollama":
		providers.Ollama.DefaultModel = model
	}
}

func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return "****" + s[len(s)-4:]
}
