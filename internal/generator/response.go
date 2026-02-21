package generator

import "time"

// Response represents the result of image generation
type Response struct {
	Images        []Image       `json:"images"`
	Model         string        `json:"model"`
	Provider      string        `json:"provider"`
	RevisedPrompt string        `json:"revised_prompt,omitempty"`
	GeneratedAt   time.Time     `json:"generated_at"`
	Duration      time.Duration `json:"duration"`
}

// Image represents a generated image
type Image struct {
	Data   []byte `json:"data,omitempty"`
	URL    string `json:"url,omitempty"`
	Format string `json:"format"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	Seed   *int64 `json:"seed,omitempty"`
	Index  int    `json:"index"`
}
