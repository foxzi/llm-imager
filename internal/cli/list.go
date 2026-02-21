package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

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

	cmd := &cobra.Command{
		Use:     "models",
		Aliases: []string{"m"},
		Short:   "List available models",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
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

			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVarP(&providerFilter, "provider", "p", "",
		"filter by provider")

	return cmd
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
