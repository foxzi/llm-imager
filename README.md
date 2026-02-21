# llm-imager

CLI tool for generating images via various AI APIs.

## Features

- Multiple providers: OpenAI, Google Gemini, Stability AI, Replicate, OpenRouter
- Unified interface for all providers
- Configuration via environment variables or YAML config file
- Support for various generation parameters

## Installation

```bash
go install github.com/piligrim/llm-imager/cmd/llm-imager@latest
```

Or build from source:

```bash
git clone https://github.com/piligrim/llm-imager.git
cd llm-imager
go build -o llm-imager ./cmd/llm-imager
```

## Configuration

### Environment Variables

```bash
export OPENAI_API_KEY="sk-..."
export GOOGLE_API_KEY="..."
export STABILITY_API_KEY="..."
export REPLICATE_API_TOKEN="..."
export OPENROUTER_API_KEY="..."
```

### Config File

Create `~/.llm-imager.yaml`:

```yaml
defaults:
  model: "openai/dall-e-3"
  size: "1024x1024"
  quality: "standard"

providers:
  openai:
    api_key: "sk-..."
    enabled: true
```

## Usage

### Basic Usage

```bash
# Generate with default model
llm-imager -p "a sunset over mountains" -o sunset.png

# Use specific model
llm-imager -m google/gemini-2.5-flash-image -p "abstract art" -o art.png

# OpenAI DALL-E 3 with HD quality
llm-imager -m openai/dall-e-3 -p "futuristic city" --quality hd --style vivid -o city.png

# Stability AI with negative prompt
llm-imager -m stability/stable-image-core -p "beautiful landscape" --negative-prompt "blurry, low quality" -o landscape.png
```

### Available Options

```
-m, --model           Model to use (e.g., google/gemini-2.5-flash-image)
-p, --prompt          Text prompt for image generation (required)
-o, --output          Output file path (required)
--size                Image size (e.g., 1024x1024)
--quality             Image quality (standard/hd or low/medium/high)
--style               Image style (natural/vivid)
-n, --count           Number of images to generate
--seed                Seed for reproducibility
--negative-prompt     Negative prompt (things to avoid)
--aspect-ratio        Aspect ratio (e.g., 16:9, 1:1)
--steps               Number of generation steps
--provider            Explicit provider selection
```

### List Providers and Models

```bash
# List all providers
llm-imager list providers

# List all models
llm-imager list models

# List models for specific provider
llm-imager list models -p openai
```

## Supported Models

### OpenAI
- `openai/dall-e-3` - DALL-E 3
- `openai/dall-e-2` - DALL-E 2
- `openai/gpt-image-1` - GPT Image 1

### Google Gemini
- `google/gemini-2.5-flash-image` - Gemini 2.5 Flash Image

### Stability AI
- `stability/stable-image-core` - Stable Image Core
- `stability/stable-image-ultra` - Stable Image Ultra
- `stability/sd3-large` - Stable Diffusion 3 Large

### Replicate
- `replicate/flux-1.1-pro` - FLUX 1.1 Pro
- `replicate/flux-schnell` - FLUX Schnell
- `replicate/sdxl` - Stable Diffusion XL

### OpenRouter
- `openrouter/google/gemini-2.5-flash-preview-image-generation`
- `openrouter/openai/gpt-image-1`

## License

MIT
