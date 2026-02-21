package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piligrim/llm-imager/internal/config"
	"github.com/piligrim/llm-imager/internal/provider"
)

var (
	cfgFile  string
	cfg      *config.Config
	registry *provider.Registry
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "llm-imager",
		Short: "CLI tool for generating images via various AI APIs",
		Long: `llm-imager is a command-line tool for generating images using
multiple AI providers including OpenAI (DALL-E), Google Gemini,
Stability AI, Replicate, and OpenRouter.

Examples:
  llm-imager -p "a sunset over mountains" -o sunset.png
  llm-imager -m google/gemini-2.5-flash-image -p "abstract art" -o art.png
  llm-imager -m openai/dall-e-3 -p "futuristic city" --quality hd -o city.png`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
		SilenceUsage: true,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default: ~/.llm-imager.yaml)")

	rootCmd.AddCommand(
		newGenerateCmd(),
		newListCmd(),
		newVersionCmd(),
	)

	return rootCmd
}

func initConfig() error {
	loader := config.NewLoader()

	var err error
	if cfgFile != "" {
		cfg, err = loader.LoadFromFile(cfgFile)
	} else {
		cfg, err = loader.Load()
	}
	if err != nil {
		return err
	}

	registry = provider.NewRegistry()
	if err := initProviders(); err != nil {
		return err
	}

	return nil
}

func initProviders() error {
	// OpenAI
	if cfg.Providers.OpenAI.Enabled {
		openai := provider.NewOpenAI(&provider.ProviderConfig{
			APIKey:     cfg.Providers.OpenAI.APIKey,
			BaseURL:    cfg.Providers.OpenAI.BaseURL,
			MaxRetries: cfg.Providers.OpenAI.MaxRetries,
		})
		registry.Register(openai)
	}

	// Google Gemini
	if cfg.Providers.Google.Enabled {
		google, err := provider.NewGoogle(&provider.ProviderConfig{
			APIKey:     cfg.Providers.Google.APIKey,
			MaxRetries: cfg.Providers.Google.MaxRetries,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to initialize Google provider: %v\n", err)
		} else {
			registry.Register(google)
		}
	}

	// OpenRouter
	if cfg.Providers.OpenRouter.Enabled {
		openrouter := provider.NewOpenRouter(&provider.ProviderConfig{
			APIKey:     cfg.Providers.OpenRouter.APIKey,
			BaseURL:    cfg.Providers.OpenRouter.BaseURL,
			MaxRetries: cfg.Providers.OpenRouter.MaxRetries,
		})
		registry.Register(openrouter)
	}

	// Stability AI
	if cfg.Providers.Stability.Enabled {
		stability := provider.NewStability(&provider.ProviderConfig{
			APIKey:     cfg.Providers.Stability.APIKey,
			MaxRetries: cfg.Providers.Stability.MaxRetries,
		})
		registry.Register(stability)
	}

	// Replicate
	if cfg.Providers.Replicate.Enabled {
		replicate := provider.NewReplicate(&provider.ProviderConfig{
			APIKey:     cfg.Providers.Replicate.APIKey,
			MaxRetries: cfg.Providers.Replicate.MaxRetries,
		})
		registry.Register(replicate)
	}

	return nil
}

// Execute runs the CLI
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
