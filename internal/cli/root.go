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
	opts := &generateOptions{}

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
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("prompt") {
				return cmd.Help()
			}
			if !cmd.Flags().Changed("output") {
				return fmt.Errorf("required flag \"output\" not set")
			}
			opts.hasSeed = cmd.Flags().Changed("seed")
			opts.hasDryRun = cmd.Flags().Changed("dry-run")
			return runGenerate(cmd.Context(), opts)
		},
		SilenceUsage: true,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default: ~/.llm-imager.yaml)")

	rootCmd.Flags().StringVarP(&opts.model, "model", "m", "",
		"model to use (e.g., google/gemini-2.5-flash-image)")
	rootCmd.Flags().StringVarP(&opts.prompt, "prompt", "p", "",
		"text prompt for image generation")
	rootCmd.Flags().StringVarP(&opts.outputPath, "output", "o", "",
		"output file path")
	rootCmd.Flags().StringVar(&opts.size, "size", "",
		"image size (e.g., 1024x1024)")
	rootCmd.Flags().StringVar(&opts.quality, "quality", "",
		"image quality (standard/hd or low/medium/high)")
	rootCmd.Flags().StringVar(&opts.style, "style", "",
		"image style (natural/vivid)")
	rootCmd.Flags().IntVarP(&opts.count, "count", "n", 0,
		"number of images to generate")
	rootCmd.Flags().Int64Var(&opts.seed, "seed", 0,
		"seed for reproducibility")
	rootCmd.Flags().StringVar(&opts.negativePrompt, "negative-prompt", "",
		"negative prompt (things to avoid)")
	rootCmd.Flags().StringVar(&opts.aspectRatio, "aspect-ratio", "",
		"aspect ratio (e.g., 16:9, 1:1)")
	rootCmd.Flags().IntVar(&opts.steps, "steps", 0,
		"number of generation steps (Stability AI, Replicate)")
	rootCmd.Flags().StringVar(&opts.providerName, "provider", "",
		"explicit provider (openai/google/stability/replicate/openrouter)")
	rootCmd.Flags().BoolVar(&opts.dryRun, "dry-run", false,
		"generate placeholder images without API calls")

	rootCmd.AddCommand(
		newGenerateCmd(),
		newListCmd(),
		newVersionCmd(),
		newCompletionCmd(),
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
