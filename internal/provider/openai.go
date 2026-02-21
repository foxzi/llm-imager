package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/piligrim/llm-imager/internal/generator"
	"github.com/piligrim/llm-imager/pkg/httputil"
)

const (
	openaiBaseURL  = "https://api.openai.com/v1"
	ModelDALLE3    = "dall-e-3"
	ModelDALLE2    = "dall-e-2"
	ModelGPTImage1 = "gpt-image-1"
)

// OpenAI implements the Provider interface for OpenAI DALL-E
type OpenAI struct {
	apiKey     string
	baseURL    string
	httpClient *httputil.Client
}

// NewOpenAI creates a new OpenAI provider
func NewOpenAI(cfg *ProviderConfig) *OpenAI {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = openaiBaseURL
	}

	return &OpenAI{
		apiKey:     cfg.APIKey,
		baseURL:    baseURL,
		httpClient: httputil.NewClient(httputil.WithRetries(cfg.MaxRetries)),
	}
}

func (o *OpenAI) Name() string {
	return "openai"
}

func (o *OpenAI) SupportedModels() []Model {
	return []Model{
		{
			ID:       "openai/dall-e-3",
			Name:     "DALL-E 3",
			Provider: "openai",
			Sizes:    []string{"1024x1024", "1792x1024", "1024x1792"},
			Features: []string{"quality", "style"},
		},
		{
			ID:       "openai/dall-e-2",
			Name:     "DALL-E 2",
			Provider: "openai",
			Sizes:    []string{"256x256", "512x512", "1024x1024"},
			Features: []string{},
		},
		{
			ID:       "openai/gpt-image-1",
			Name:     "GPT Image 1",
			Provider: "openai",
			Sizes:    []string{"1024x1024", "1024x1536", "1536x1024"},
			Features: []string{"quality"},
		},
	}
}

func (o *OpenAI) ValidateRequest(req *generator.Request) error {
	if o.apiKey == "" {
		return fmt.Errorf("OpenAI API key is required (set OPENAI_API_KEY)")
	}

	model := o.extractModelName(req.Model)

	// Validate sizes for DALL-E 3
	if model == ModelDALLE3 && req.Size != "" {
		validSizes := map[string]bool{
			"1024x1024": true, "1792x1024": true, "1024x1792": true,
		}
		if !validSizes[req.Size] {
			return fmt.Errorf("invalid size %s for DALL-E 3", req.Size)
		}
	}

	// Validate quality
	if req.Quality != "" {
		validQuality := map[string]bool{
			"standard": true, "hd": true, "low": true, "medium": true, "high": true,
		}
		if !validQuality[req.Quality] {
			return fmt.Errorf("invalid quality %s", req.Quality)
		}
	}

	// Validate style
	if req.Style != "" && req.Style != "natural" && req.Style != "vivid" {
		return fmt.Errorf("invalid style %s, valid: natural, vivid", req.Style)
	}

	return nil
}

type openaiImageRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	Quality        string `json:"quality,omitempty"`
	Style          string `json:"style,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
}

type openaiImageResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL           string `json:"url,omitempty"`
		B64JSON       string `json:"b64_json,omitempty"`
		RevisedPrompt string `json:"revised_prompt,omitempty"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

func (o *OpenAI) Generate(ctx context.Context, req *generator.Request) (*generator.Response, error) {
	if err := o.ValidateRequest(req); err != nil {
		return nil, err
	}

	startTime := time.Now()

	count := req.Count
	if count <= 0 {
		count = 1
	}

	model := o.extractModelName(req.Model)

	apiReq := openaiImageRequest{
		Model:          model,
		Prompt:         req.Prompt,
		N:              count,
		Size:           req.Size,
		Quality:        req.Quality,
		Style:          req.Style,
		ResponseFormat: "b64_json",
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		o.baseURL+"/images/generations",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiResp openaiImageResponse
		json.Unmarshal(respBody, &apiResp)
		if apiResp.Error != nil {
			return nil, fmt.Errorf("OpenAI API error: %s", apiResp.Error.Message)
		}
		return nil, fmt.Errorf("OpenAI API error: status %d", resp.StatusCode)
	}

	var apiResp openaiImageResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	images := make([]generator.Image, 0, len(apiResp.Data))
	for i, img := range apiResp.Data {
		var data []byte
		if img.B64JSON != "" {
			data, err = base64.StdEncoding.DecodeString(img.B64JSON)
			if err != nil {
				return nil, fmt.Errorf("failed to decode image: %w", err)
			}
		}

		images = append(images, generator.Image{
			Data:   data,
			URL:    img.URL,
			Format: "png",
			Index:  i,
		})
	}

	revisedPrompt := ""
	if len(apiResp.Data) > 0 {
		revisedPrompt = apiResp.Data[0].RevisedPrompt
	}

	return &generator.Response{
		Images:        images,
		Model:         req.Model,
		Provider:      o.Name(),
		RevisedPrompt: revisedPrompt,
		GeneratedAt:   time.Now(),
		Duration:      time.Since(startTime),
	}, nil
}

func (o *OpenAI) extractModelName(model string) string {
	if strings.HasPrefix(model, "openai/") {
		return strings.TrimPrefix(model, "openai/")
	}
	return model
}
