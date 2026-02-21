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

const openrouterBaseURL = "https://openrouter.ai/api/v1"

// OpenRouter implements the Provider interface for OpenRouter
type OpenRouter struct {
	apiKey     string
	baseURL    string
	httpClient *httputil.Client
}

// NewOpenRouter creates a new OpenRouter provider
func NewOpenRouter(cfg *ProviderConfig) *OpenRouter {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = openrouterBaseURL
	}

	return &OpenRouter{
		apiKey:     cfg.APIKey,
		baseURL:    baseURL,
		httpClient: httputil.NewClient(httputil.WithRetries(cfg.MaxRetries)),
	}
}

func (o *OpenRouter) Name() string {
	return "openrouter"
}

func (o *OpenRouter) SupportedModels() []Model {
	return []Model{
		{
			ID:       "openrouter/google/gemini-2.5-flash-image",
			Name:     "Gemini 2.5 Flash Image (via OpenRouter)",
			Provider: "openrouter",
			Features: []string{"aspect_ratio", "image_size"},
		},
		{
			ID:       "openrouter/google/gemini-3-pro-image-preview",
			Name:     "Gemini 3 Pro Image Preview (via OpenRouter)",
			Provider: "openrouter",
			Features: []string{"aspect_ratio", "image_size"},
		},
		{
			ID:       "openrouter/openai/gpt-5-image",
			Name:     "GPT-5 Image (via OpenRouter)",
			Provider: "openrouter",
			Features: []string{"aspect_ratio"},
		},
		{
			ID:       "openrouter/openai/gpt-5-image-mini",
			Name:     "GPT-5 Image Mini (via OpenRouter)",
			Provider: "openrouter",
			Features: []string{"aspect_ratio"},
		},
	}
}

func (o *OpenRouter) ValidateRequest(req *generator.Request) error {
	if o.apiKey == "" {
		return fmt.Errorf("OpenRouter API key is required (set OPENROUTER_API_KEY)")
	}
	return nil
}

type openrouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openrouterImageConfig struct {
	AspectRatio string `json:"aspect_ratio,omitempty"`
	ImageSize   string `json:"image_size,omitempty"`
}

type openrouterRequest struct {
	Model       string              `json:"model"`
	Messages    []openrouterMessage `json:"messages"`
	Modalities  []string            `json:"modalities"`
	ImageConfig *openrouterImageConfig `json:"image_config,omitempty"`
}

type openrouterResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content any    `json:"content"` // Can be string or array
			Images  []struct {
				ImageURL struct {
					URL string `json:"url"`
				} `json:"image_url"`
			} `json:"images,omitempty"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

func (o *OpenRouter) Generate(ctx context.Context, req *generator.Request) (*generator.Response, error) {
	if err := o.ValidateRequest(req); err != nil {
		return nil, err
	}

	startTime := time.Now()

	model := o.extractModelName(req.Model)

	apiReq := openrouterRequest{
		Model: model,
		Messages: []openrouterMessage{
			{Role: "user", Content: req.Prompt},
		},
		Modalities: []string{"image", "text"},
	}

	if req.AspectRatio != "" || req.Size != "" {
		apiReq.ImageConfig = &openrouterImageConfig{
			AspectRatio: req.AspectRatio,
			ImageSize:   o.mapImageSize(req.Size),
		}
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		o.baseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("HTTP-Referer", "https://github.com/piligrim/llm-imager")

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
		var apiResp openrouterResponse
		json.Unmarshal(respBody, &apiResp)
		if apiResp.Error != nil {
			return nil, fmt.Errorf("OpenRouter API error: %s", apiResp.Error.Message)
		}
		return nil, fmt.Errorf("OpenRouter API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var apiResp openrouterResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	images := make([]generator.Image, 0)

	if len(apiResp.Choices) > 0 {
		msg := apiResp.Choices[0].Message

		// Check images array
		for i, img := range msg.Images {
			if img.ImageURL.URL != "" {
				url := img.ImageURL.URL
				var imageData []byte
				var format string
				var err error

				// Check if it's a data URL
				if strings.HasPrefix(url, "data:image/") {
					imageData, format, err = o.parseDataURL(url)
				} else {
					imageData, format, err = o.downloadImage(ctx, url)
				}

				if err != nil {
					return nil, fmt.Errorf("failed to get image: %w", err)
				}
				images = append(images, generator.Image{
					Data:   imageData,
					Format: format,
					Index:  i,
				})
			}
		}

		// Check content array for base64 images
		if content, ok := msg.Content.([]any); ok {
			for i, item := range content {
				if m, ok := item.(map[string]any); ok {
					if m["type"] == "image" {
						if imgData, ok := m["image"].(map[string]any); ok {
							if url, ok := imgData["url"].(string); ok {
								// Check if it's a data URL
								if strings.HasPrefix(url, "data:image/") {
									data, format, err := o.parseDataURL(url)
									if err == nil {
										images = append(images, generator.Image{
											Data:   data,
											Format: format,
											Index:  i,
										})
									}
								} else {
									imageData, format, err := o.downloadImage(ctx, url)
									if err == nil {
										images = append(images, generator.Image{
											Data:   imageData,
											URL:    url,
											Format: format,
											Index:  i,
										})
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no images in response")
	}

	return &generator.Response{
		Images:      images,
		Model:       req.Model,
		Provider:    o.Name(),
		GeneratedAt: time.Now(),
		Duration:    time.Since(startTime),
	}, nil
}

func (o *OpenRouter) extractModelName(model string) string {
	if name, found := strings.CutPrefix(model, "openrouter/"); found {
		return name
	}
	return model
}

func (o *OpenRouter) mapImageSize(size string) string {
	switch size {
	case "1024x1024", "1K":
		return "1K"
	case "2048x2048", "2K":
		return "2K"
	case "4096x4096", "4K":
		return "4K"
	default:
		return ""
	}
}

func (o *OpenRouter) downloadImage(ctx context.Context, url string) ([]byte, string, error) {
	resp, err := o.httpClient.Get(ctx, url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	format := "png"
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "jpeg") {
		format = "jpeg"
	} else if strings.Contains(contentType, "webp") {
		format = "webp"
	}

	return data, format, nil
}

func (o *OpenRouter) parseDataURL(dataURL string) ([]byte, string, error) {
	// Format: data:image/png;base64,<data>
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid data URL")
	}

	format := "png"
	if strings.Contains(parts[0], "jpeg") {
		format = "jpeg"
	} else if strings.Contains(parts[0], "webp") {
		format = "webp"
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, "", err
	}

	return data, format, nil
}
