package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/piligrim/llm-imager/internal/provider"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available providers and models",
		Long:  `Display information about available providers, models, and their capabilities.`,
	}

	cmd.AddCommand(
		newListProvidersCmd(),
		newListModelsCmd(),
	)

	return cmd
}

func newListProvidersCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "providers",
		Aliases: []string{"p"},
		Short:   "List available providers",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "PROVIDER\tSTATUS\tMODELS")

			for _, p := range registry.ListProviders() {
				status := "ready"
				if err := checkProviderAPIKey(p.Name()); err != nil {
					status = "no api key"
				}
				fmt.Fprintf(w, "%s\t%s\t%d\n", p.Name(), status, len(p.SupportedModels()))
			}

			w.Flush()
			return nil
		},
	}
}

func newListModelsCmd() *cobra.Command {
	var providerFilter string
	var showPrices bool

	cmd := &cobra.Command{
		Use:     "models",
		Aliases: []string{"m"},
		Short:   "List available models",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			if showPrices {
				fmt.Fprintln(w, "MODEL\tPROVIDER\tPRICE (per 1M tokens)")

				// Fetch models with prices from OpenRouter
				models, err := provider.FetchImageModels(context.Background())
				if err != nil {
					return fmt.Errorf("failed to fetch prices: %w", err)
				}

				for _, model := range models {
					if providerFilter != "" && model.Provider != providerFilter {
						continue
					}
					price := formatPrice(model.Pricing)
					fmt.Fprintf(w, "%s\t%s\t%s\n", model.ID, model.Provider, price)
				}
			} else {
				fmt.Fprintln(w, "MODEL\tPROVIDER\tFEATURES")

				for _, model := range registry.ListModels() {
					if providerFilter != "" && model.Provider != providerFilter {
						continue
					}
					features := strings.Join(model.Features, ", ")
					if features == "" {
						features = "-"
					}
					fmt.Fprintf(w, "%s\t%s\t%s\n", model.ID, model.Provider, features)
				}
			}

			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVarP(&providerFilter, "provider", "p", "",
		"filter by provider")
	cmd.Flags().BoolVar(&showPrices, "prices", false,
		"show pricing info from OpenRouter API")

	return cmd
}

func formatPrice(p *provider.Pricing) string {
	if p == nil {
		return "-"
	}

	prompt, err := strconv.ParseFloat(p.Prompt, 64)
	if err != nil || prompt < 0 {
		return "-"
	}
	completion, err := strconv.ParseFloat(p.Completion, 64)
	if err != nil || completion < 0 {
		return "-"
	}

	// Convert to per 1M tokens
	promptPerM := prompt * 1_000_000
	completionPerM := completion * 1_000_000

	return fmt.Sprintf("$%.2f / $%.2f", promptPerM, completionPerM)
}

func checkProviderAPIKey(name string) error {
	switch name {
	case "openai":
		if cfg.Providers.OpenAI.APIKey == "" {
			return fmt.Errorf("no API key")
		}
	case "google":
		if cfg.Providers.Google.APIKey == "" {
			return fmt.Errorf("no API key")
		}
	case "openrouter":
		if cfg.Providers.OpenRouter.APIKey == "" {
			return fmt.Errorf("no API key")
		}
	case "stability":
		if cfg.Providers.Stability.APIKey == "" {
			return fmt.Errorf("no API key")
		}
	case "replicate":
		if cfg.Providers.Replicate.APIKey == "" {
			return fmt.Errorf("no API key")
		}
	}
	return nil
}
