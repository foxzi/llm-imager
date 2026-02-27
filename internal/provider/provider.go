package provider

import (
	"context"

	"github.com/piligrim/llm-imager/internal/generator"
)

// Provider defines the contract for all image generation providers
type Provider interface {
	// Name returns the unique provider name
	Name() string

	// Generate performs image generation
	Generate(ctx context.Context, req *generator.Request) (*generator.Response, error)

	// SupportedModels returns the list of supported models
	SupportedModels() []Model

	// ValidateRequest checks request compatibility with the provider
	ValidateRequest(req *generator.Request) error
}

// Model describes an image generation model
type Model struct {
	ID       string   // e.g., google/gemini-2.5-flash-image
	Name     string   // Human-readable name
	Provider string   // Provider name
	Sizes    []string // Supported sizes
	Features []string // negative_prompt, seed, steps, etc.
	Pricing  *Pricing // Pricing info (optional)
}

// Pricing contains model pricing information (per token)
type Pricing struct {
	Prompt     string // Price per prompt token
	Completion string // Price per completion token
}

// ProviderConfig contains provider configuration
type ProviderConfig struct {
	APIKey     string
	BaseURL    string
	MaxRetries int
}
