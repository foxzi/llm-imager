package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/piligrim/llm-imager/internal/generator"
	"github.com/piligrim/llm-imager/internal/output"
	"github.com/piligrim/llm-imager/internal/provider"
)

type generateOptions struct {
	model          string
	prompt         string
	outputPath     string
	size           string
	quality        string
	style          string
	count          int
	seed           int64
	hasSeed        bool
	negativePrompt string
	aspectRatio    string
	steps          int
	providerName   string
	dryRun         bool
	hasDryRun      bool
}

func newGenerateCmd() *cobra.Command {
	opts := &generateOptions{}

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate images from text prompt",
		Long: `Generate images using AI models from various providers.

The model can be specified in the format "provider/model" (e.g., "google/gemini-2.5-flash-image")
or just the model name if the provider can be auto-detected.`,
		Aliases: []string{"gen", "g"},
		Example: `  llm-imager generate -p "a beautiful landscape" -o landscape.png
  llm-imager g -m openai/dall-e-3 -p "abstract art" -o art.png
  llm-imager generate -m stability/stable-image-core -p "cyberpunk city" --negative-prompt "blurry" -o city.png`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.hasSeed = cmd.Flags().Changed("seed")
			opts.hasDryRun = cmd.Flags().Changed("dry-run")
			return runGenerate(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.model, "model", "m", "",
		"model to use (e.g., google/gemini-2.5-flash-image)")
	cmd.Flags().StringVarP(&opts.prompt, "prompt", "p", "",
		"text prompt for image generation (required)")
	cmd.Flags().StringVarP(&opts.outputPath, "output", "o", "",
		"output file path (required)")
	cmd.Flags().StringVar(&opts.size, "size", "",
		"image size (e.g., 1024x1024)")
	cmd.Flags().StringVar(&opts.quality, "quality", "",
		"image quality (standard/hd or low/medium/high)")
	cmd.Flags().StringVar(&opts.style, "style", "",
		"image style (natural/vivid)")
	cmd.Flags().IntVarP(&opts.count, "count", "n", 0,
		"number of images to generate")
	cmd.Flags().Int64Var(&opts.seed, "seed", 0,
		"seed for reproducibility")
	cmd.Flags().StringVar(&opts.negativePrompt, "negative-prompt", "",
		"negative prompt (things to avoid)")
	cmd.Flags().StringVar(&opts.aspectRatio, "aspect-ratio", "",
		"aspect ratio (e.g., 16:9, 1:1)")
	cmd.Flags().IntVar(&opts.steps, "steps", 0,
		"number of generation steps (Stability AI, Replicate)")
	cmd.Flags().StringVar(&opts.providerName, "provider", "",
		"explicit provider (openai/google/stability/replicate/openrouter)")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false,
		"generate placeholder images without API calls")

	cmd.MarkFlagRequired("prompt")
	cmd.MarkFlagRequired("output")

	return cmd
}

func runGenerate(ctx context.Context, opts *generateOptions) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	applyDefaults(opts)

	var seedPtr *int64
	if opts.hasSeed {
		seedPtr = &opts.seed
	}

	req := &generator.Request{
		Model:          opts.model,
		Prompt:         opts.prompt,
		Size:           opts.size,
		Quality:        opts.quality,
		Style:          opts.style,
		Count:          opts.count,
		Seed:           seedPtr,
		NegativePrompt: opts.negativePrompt,
		AspectRatio:    opts.aspectRatio,
		Steps:          opts.steps,
	}

	var p interface {
		Name() string
		Generate(context.Context, *generator.Request) (*generator.Response, error)
		ValidateRequest(*generator.Request) error
	}
	var err error

	if opts.dryRun {
		p = provider.NewDryRun()
		fmt.Printf("Dry-run mode: generating placeholder image (%s)...\n", opts.size)
	} else {
		if opts.providerName != "" {
			p, err = registry.GetByName(opts.providerName)
		} else {
			p, err = registry.GetByModel(opts.model)
		}
		if err != nil {
			return fmt.Errorf("failed to get provider: %w", err)
		}
		fmt.Printf("Generating image with %s using model %s...\n", p.Name(), opts.model)
	}

	resp, err := p.Generate(ctx, req)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	writer := output.NewWriter(cfg.Output.Format)
	paths, err := writer.Write(resp.Images, opts.outputPath)
	if err != nil {
		return fmt.Errorf("failed to save images: %w", err)
	}

	for _, path := range paths {
		fmt.Printf("Saved: %s\n", path)
	}

	fmt.Printf("Generation completed in %s\n", resp.Duration.Round(100*1e6))

	return nil
}

func applyDefaults(opts *generateOptions) {
	if opts.model == "" {
		opts.model = cfg.Defaults.Model
	}
	if opts.size == "" && cfg.Defaults.Size != "" {
		opts.size = cfg.Defaults.Size
	}
	if opts.quality == "" && cfg.Defaults.Quality != "" {
		opts.quality = cfg.Defaults.Quality
	}
	if opts.style == "" && cfg.Defaults.Style != "" {
		opts.style = cfg.Defaults.Style
	}
	if opts.count == 0 {
		opts.count = cfg.Defaults.Count
	}
	if opts.aspectRatio == "" && cfg.Defaults.AspectRatio != "" {
		opts.aspectRatio = cfg.Defaults.AspectRatio
	}
	if !opts.hasDryRun && cfg.Defaults.DryRun {
		opts.dryRun = true
	}
}
