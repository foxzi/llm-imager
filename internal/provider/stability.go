package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/piligrim/llm-imager/internal/generator"
	"github.com/piligrim/llm-imager/pkg/httputil"
)

const stabilityBaseURL = "https://api.stability.ai"

// Stability implements the Provider interface for Stability AI
type Stability struct {
	apiKey     string
	baseURL    string
	httpClient *httputil.Client
}

// NewStability creates a new Stability AI provider
func NewStability(cfg *ProviderConfig) *Stability {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = stabilityBaseURL
	}

	return &Stability{
		apiKey:     cfg.APIKey,
		baseURL:    baseURL,
		httpClient: httputil.NewClient(httputil.WithRetries(cfg.MaxRetries)),
	}
}

func (s *Stability) Name() string {
	return "stability"
}

func (s *Stability) SupportedModels() []Model {
	return []Model{
		{
			ID:       "stability/stable-image-core",
			Name:     "Stable Image Core",
			Provider: "stability",
			Sizes:    []string{"1024x1024", "1152x896", "896x1152"},
			Features: []string{"negative_prompt", "seed", "aspect_ratio", "style_preset"},
		},
		{
			ID:       "stability/stable-image-ultra",
			Name:     "Stable Image Ultra",
			Provider: "stability",
			Sizes:    []string{"1024x1024"},
			Features: []string{"negative_prompt", "seed", "aspect_ratio"},
		},
		{
			ID:       "stability/sd3-large",
			Name:     "Stable Diffusion 3 Large",
			Provider: "stability",
			Sizes:    []string{"1024x1024"},
			Features: []string{"negative_prompt", "seed"},
		},
	}
}

func (s *Stability) ValidateRequest(req *generator.Request) error {
	if s.apiKey == "" {
		return fmt.Errorf("Stability API key is required (set STABILITY_API_KEY)")
	}
	return nil
}

func (s *Stability) Generate(ctx context.Context, req *generator.Request) (*generator.Response, error) {
	if err := s.ValidateRequest(req); err != nil {
		return nil, err
	}

	startTime := time.Now()

	model := s.extractModelName(req.Model)
	endpoint := s.getEndpoint(model)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	writer.WriteField("prompt", req.Prompt)

	if req.NegativePrompt != "" {
		writer.WriteField("negative_prompt", req.NegativePrompt)
	}

	if req.AspectRatio != "" {
		writer.WriteField("aspect_ratio", req.AspectRatio)
	}

	if req.Seed != nil {
		writer.WriteField("seed", fmt.Sprintf("%d", *req.Seed))
	}

	writer.WriteField("output_format", "png")

	writer.Close()

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		s.baseURL+endpoint,
		&body,
	)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Accept", "image/*")

	resp, err := s.httpClient.Do(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Name    string `json:"name"`
			Message string `json:"message"`
		}
		json.Unmarshal(respBody, &errResp)
		if errResp.Message != "" {
			return nil, fmt.Errorf("Stability API error: %s", errResp.Message)
		}
		return nil, fmt.Errorf("Stability API error: status %d", resp.StatusCode)
	}

	format := "png"
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "jpeg") {
		format = "jpeg"
	} else if strings.Contains(contentType, "webp") {
		format = "webp"
	}

	images := []generator.Image{
		{
			Data:   respBody,
			Format: format,
			Index:  0,
		},
	}

	return &generator.Response{
		Images:      images,
		Model:       req.Model,
		Provider:    s.Name(),
		GeneratedAt: time.Now(),
		Duration:    time.Since(startTime),
	}, nil
}

func (s *Stability) extractModelName(model string) string {
	if strings.HasPrefix(model, "stability/") {
		return strings.TrimPrefix(model, "stability/")
	}
	return model
}

func (s *Stability) getEndpoint(model string) string {
	switch model {
	case "stable-image-ultra":
		return "/v2beta/stable-image/generate/ultra"
	case "sd3-large", "sd3-large-turbo":
		return "/v2beta/stable-image/generate/sd3"
	default:
		return "/v2beta/stable-image/generate/core"
	}
}
