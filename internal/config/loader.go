package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Loader loads configuration from file and environment variables
type Loader struct {
	v *viper.Viper
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	v := viper.New()

	v.SetConfigName(".llm-imager")
	v.SetConfigType("yaml")

	// Search paths (in order of priority)
	v.AddConfigPath(".")
	if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(home)
	}
	v.AddConfigPath("/etc/llm-imager")

	// Environment variables
	v.SetEnvPrefix("LLMIMAGER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	bindEnvVariables(v)

	return &Loader{v: v}
}

func bindEnvVariables(v *viper.Viper) {
	// API keys (standard names for compatibility)
	v.BindEnv("providers.openai.api_key", "OPENAI_API_KEY")
	v.BindEnv("providers.google.api_key", "GOOGLE_API_KEY", "GEMINI_API_KEY")
	v.BindEnv("providers.stability.api_key", "STABILITY_API_KEY")
	v.BindEnv("providers.replicate.api_key", "REPLICATE_API_TOKEN")
	v.BindEnv("providers.openrouter.api_key", "OPENROUTER_API_KEY")

	// Base URLs for proxy/custom endpoints
	v.BindEnv("providers.openai.base_url", "OPENAI_BASE_URL")
	v.BindEnv("providers.openrouter.base_url", "OPENROUTER_BASE_URL")
}

// Load loads configuration from file and environment
func (l *Loader) Load() (*Config, error) {
	setDefaults(l.v)

	// Try to read config file (optional)
	if err := l.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := l.v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadFromFile loads configuration from specified file
func (l *Loader) LoadFromFile(path string) (*Config, error) {
	l.v.SetConfigFile(path)
	return l.Load()
}

// ConfigFilePath returns the path to the config file if found
func (l *Loader) ConfigFilePath() string {
	return l.v.ConfigFileUsed()
}

// DefaultConfigPath returns the default config file path
func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".llm-imager.yaml")
}

func setDefaults(v *viper.Viper) {
	// Defaults
	v.SetDefault("defaults.model", "openai/dall-e-3")
	v.SetDefault("defaults.size", "1024x1024")
	v.SetDefault("defaults.quality", "standard")
	v.SetDefault("defaults.style", "natural")
	v.SetDefault("defaults.count", 1)
	v.SetDefault("defaults.aspect_ratio", "1:1")

	// Provider defaults
	v.SetDefault("providers.openai.timeout", 60*time.Second)
	v.SetDefault("providers.openai.max_retries", 3)
	v.SetDefault("providers.openai.enabled", true)

	v.SetDefault("providers.google.timeout", 60*time.Second)
	v.SetDefault("providers.google.max_retries", 3)
	v.SetDefault("providers.google.enabled", true)

	v.SetDefault("providers.stability.timeout", 120*time.Second)
	v.SetDefault("providers.stability.max_retries", 3)
	v.SetDefault("providers.stability.enabled", true)

	v.SetDefault("providers.replicate.timeout", 300*time.Second)
	v.SetDefault("providers.replicate.max_retries", 3)
	v.SetDefault("providers.replicate.enabled", true)

	v.SetDefault("providers.openrouter.timeout", 120*time.Second)
	v.SetDefault("providers.openrouter.max_retries", 3)
	v.SetDefault("providers.openrouter.enabled", true)

	// Output
	v.SetDefault("output.directory", "./")
	v.SetDefault("output.format", "png")
}
