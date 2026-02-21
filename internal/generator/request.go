package generator

// Request represents an image generation request
type Request struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	Size           string `json:"size,omitempty"`
	Quality        string `json:"quality,omitempty"`
	Style          string `json:"style,omitempty"`
	Count          int    `json:"count,omitempty"`
	Seed           *int64 `json:"seed,omitempty"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	AspectRatio    string `json:"aspect_ratio,omitempty"`
	Steps          int    `json:"steps,omitempty"`
}
