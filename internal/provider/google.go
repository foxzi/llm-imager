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

const googleBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// Google implements the Provider interface for Google Gemini
type Google struct {
	apiKey     string
	baseURL    string
	httpClient *httputil.Client
}

// NewGoogle creates a new Google Gemini provider
func NewGoogle(cfg *ProviderConfig) (*Google, error) {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = googleBaseURL
	}

	return &Google{
		apiKey:     cfg.APIKey,
		baseURL:    baseURL,
		httpClient: httputil.NewClient(httputil.WithRetries(cfg.MaxRetries)),
	}, nil
}

func (g *Google) Name() string {
	return "google"
}

func (g *Google) SupportedModels() []Model {
	return []Model{
		{
			ID:       "google/gemini-2.0-flash-exp-image",
			Name:     "Gemini 2.0 Flash Exp Image",
			Provider: "google",
			Sizes:    []string{"1024x1024"},
			Features: []string{},
		},
		{
			ID:       "google/imagen-3.0-generate-002",
			Name:     "Imagen 3.0",
			Provider: "google",
			Sizes:    []string{"1024x1024"},
			Features: []string{"aspect_ratio"},
		},
	}
}

func (g *Google) ValidateRequest(req *generator.Request) error {
	if g.apiKey == "" {
		return fmt.Errorf("Google API key is required (set GOOGLE_API_KEY or GEMINI_API_KEY)")
	}
	return nil
}

type geminiPart struct {
	Text       string          `json:"text,omitempty"`
	InlineData *geminiDataBlob `json:"inlineData,omitempty"`
}

type geminiDataBlob struct {
	MIMEType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type geminiRequest struct {
	Contents         []geminiContent   `json:"contents"`
	GenerationConfig *geminiGenConfig  `json:"generationConfig,omitempty"`
}

type geminiGenConfig struct {
	ResponseModalities []string `json:"responseModalities,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text       string `json:"text,omitempty"`
				InlineData *struct {
					MIMEType string `json:"mimeType"`
					Data     string `json:"data"`
				} `json:"inlineData,omitempty"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func (g *Google) Generate(ctx context.Context, req *generator.Request) (*generator.Response, error) {
	if err := g.ValidateRequest(req); err != nil {
		return nil, err
	}

	startTime := time.Now()

	model := g.extractModelName(req.Model)

	apiReq := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: req.Prompt},
				},
			},
		},
		GenerationConfig: &geminiGenConfig{
			ResponseModalities: []string{"TEXT", "IMAGE"},
		},
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", g.baseURL, model, g.apiKey)

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp geminiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("Gemini API error: %s", apiResp.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Gemini API error: status %d", resp.StatusCode)
	}

	images := make([]generator.Image, 0)

	if len(apiResp.Candidates) > 0 {
		for i, part := range apiResp.Candidates[0].Content.Parts {
			if part.InlineData != nil && part.InlineData.Data != "" {
				data, err := decodeBase64(part.InlineData.Data)
				if err != nil {
					continue
				}

				format := "png"
				if strings.Contains(part.InlineData.MIMEType, "jpeg") {
					format = "jpeg"
				} else if strings.Contains(part.InlineData.MIMEType, "webp") {
					format = "webp"
				}

				images = append(images, generator.Image{
					Data:   data,
					Format: format,
					Index:  i,
				})
			}
		}
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no images generated")
	}

	return &generator.Response{
		Images:      images,
		Model:       req.Model,
		Provider:    g.Name(),
		GeneratedAt: time.Now(),
		Duration:    time.Since(startTime),
	}, nil
}

func (g *Google) extractModelName(model string) string {
	if strings.HasPrefix(model, "google/") {
		return strings.TrimPrefix(model, "google/")
	}
	return model
}

func decodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
