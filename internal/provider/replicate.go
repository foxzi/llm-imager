package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/piligrim/llm-imager/internal/generator"
	"github.com/piligrim/llm-imager/pkg/httputil"
)

const replicateBaseURL = "https://api.replicate.com/v1"

// Replicate implements the Provider interface for Replicate
type Replicate struct {
	apiKey     string
	baseURL    string
	httpClient *httputil.Client
}

// NewReplicate creates a new Replicate provider
func NewReplicate(cfg *ProviderConfig) *Replicate {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = replicateBaseURL
	}

	return &Replicate{
		apiKey:     cfg.APIKey,
		baseURL:    baseURL,
		httpClient: httputil.NewClient(httputil.WithRetries(cfg.MaxRetries)),
	}
}

func (r *Replicate) Name() string {
	return "replicate"
}

func (r *Replicate) SupportedModels() []Model {
	return []Model{
		{
			ID:       "replicate/flux-1.1-pro",
			Name:     "FLUX 1.1 Pro",
			Provider: "replicate",
			Features: []string{"aspect_ratio", "seed"},
		},
		{
			ID:       "replicate/flux-schnell",
			Name:     "FLUX Schnell",
			Provider: "replicate",
			Features: []string{"aspect_ratio", "seed"},
		},
		{
			ID:       "replicate/sdxl",
			Name:     "Stable Diffusion XL",
			Provider: "replicate",
			Features: []string{"negative_prompt", "seed", "steps"},
		},
	}
}

func (r *Replicate) ValidateRequest(req *generator.Request) error {
	if r.apiKey == "" {
		return fmt.Errorf("Replicate API token is required (set REPLICATE_API_TOKEN)")
	}
	return nil
}

type replicatePrediction struct {
	ID     string   `json:"id"`
	Status string   `json:"status"`
	Output any      `json:"output"` // Can be string or []string
	Error  string   `json:"error,omitempty"`
	URLs   struct {
		Get string `json:"get"`
	} `json:"urls"`
}

type replicateRequest struct {
	Version string         `json:"version,omitempty"`
	Model   string         `json:"model,omitempty"`
	Input   map[string]any `json:"input"`
}

func (r *Replicate) Generate(ctx context.Context, req *generator.Request) (*generator.Response, error) {
	if err := r.ValidateRequest(req); err != nil {
		return nil, err
	}

	startTime := time.Now()

	model := r.extractModelName(req.Model)
	modelRef := r.getModelRef(model)

	input := map[string]any{
		"prompt": req.Prompt,
	}

	if req.NegativePrompt != "" {
		input["negative_prompt"] = req.NegativePrompt
	}

	if req.AspectRatio != "" {
		input["aspect_ratio"] = req.AspectRatio
	}

	if req.Seed != nil {
		input["seed"] = *req.Seed
	}

	if req.Steps > 0 {
		input["num_inference_steps"] = req.Steps
	}

	apiReq := replicateRequest{
		Model: modelRef,
		Input: input,
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		r.baseURL+"/predictions",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+r.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Prefer", "wait")

	resp, err := r.httpClient.Do(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Replicate API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var prediction replicatePrediction
	if err := json.Unmarshal(respBody, &prediction); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Poll if not completed
	for prediction.Status != "succeeded" && prediction.Status != "failed" {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}

		prediction, err = r.getPrediction(ctx, prediction.URLs.Get)
		if err != nil {
			return nil, err
		}
	}

	if prediction.Status == "failed" {
		return nil, fmt.Errorf("Replicate generation failed: %s", prediction.Error)
	}

	// Extract image URLs
	var imageURLs []string
	switch output := prediction.Output.(type) {
	case string:
		imageURLs = []string{output}
	case []any:
		for _, u := range output {
			if s, ok := u.(string); ok {
				imageURLs = append(imageURLs, s)
			}
		}
	}

	if len(imageURLs) == 0 {
		return nil, fmt.Errorf("no images in response")
	}

	images := make([]generator.Image, 0, len(imageURLs))
	for i, url := range imageURLs {
		data, format, err := r.downloadImage(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("failed to download image: %w", err)
		}
		images = append(images, generator.Image{
			Data:   data,
			URL:    url,
			Format: format,
			Index:  i,
		})
	}

	return &generator.Response{
		Images:      images,
		Model:       req.Model,
		Provider:    r.Name(),
		GeneratedAt: time.Now(),
		Duration:    time.Since(startTime),
	}, nil
}

func (r *Replicate) getPrediction(ctx context.Context, url string) (replicatePrediction, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return replicatePrediction{}, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+r.apiKey)

	resp, err := r.httpClient.Do(ctx, httpReq)
	if err != nil {
		return replicatePrediction{}, err
	}
	defer resp.Body.Close()

	var prediction replicatePrediction
	if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
		return replicatePrediction{}, err
	}

	return prediction, nil
}

func (r *Replicate) downloadImage(ctx context.Context, url string) ([]byte, string, error) {
	resp, err := r.httpClient.Get(ctx, url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	format := "webp"
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "png") {
		format = "png"
	} else if strings.Contains(contentType, "jpeg") {
		format = "jpeg"
	}

	return data, format, nil
}

func (r *Replicate) extractModelName(model string) string {
	if strings.HasPrefix(model, "replicate/") {
		return strings.TrimPrefix(model, "replicate/")
	}
	return model
}

func (r *Replicate) getModelRef(model string) string {
	switch model {
	case "flux-1.1-pro":
		return "black-forest-labs/flux-1.1-pro"
	case "flux-schnell":
		return "black-forest-labs/flux-schnell"
	case "sdxl":
		return "stability-ai/sdxl"
	default:
		return model
	}
}
