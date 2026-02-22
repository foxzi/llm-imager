package config

import "time"

// Config is the root configuration structure
type Config struct {
	Defaults  DefaultsConfig  `mapstructure:"defaults"`
	Providers ProvidersConfig `mapstructure:"providers"`
	Output    OutputConfig    `mapstructure:"output"`
}

// DefaultsConfig contains default generation settings
type DefaultsConfig struct {
	Model       string `mapstructure:"model"`
	Provider    string `mapstructure:"provider"`
	Size        string `mapstructure:"size"`
	Quality     string `mapstructure:"quality"`
	Style       string `mapstructure:"style"`
	Count       int    `mapstructure:"count"`
	AspectRatio string `mapstructure:"aspect_ratio"`
	DryRun      bool   `mapstructure:"dry_run"`
}

// ProvidersConfig contains settings for all providers
type ProvidersConfig struct {
	OpenAI     ProviderSettings `mapstructure:"openai"`
	Google     ProviderSettings `mapstructure:"google"`
	Stability  ProviderSettings `mapstructure:"stability"`
	Replicate  ProviderSettings `mapstructure:"replicate"`
	OpenRouter ProviderSettings `mapstructure:"openrouter"`
}

// ProviderSettings contains settings for a single provider
type ProviderSettings struct {
	APIKey     string        `mapstructure:"api_key"`
	BaseURL    string        `mapstructure:"base_url"`
	Timeout    time.Duration `mapstructure:"timeout"`
	MaxRetries int           `mapstructure:"max_retries"`
	Enabled    bool          `mapstructure:"enabled"`
}

// OutputConfig contains output settings
type OutputConfig struct {
	Directory string `mapstructure:"directory"`
	Format    string `mapstructure:"format"`
}
