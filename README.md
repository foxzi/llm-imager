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
- `openrouter/google/gemini-2.5-flash-image` - Gemini 2.5 Flash Image
- `openrouter/google/gemini-3-pro-image-preview` - Gemini 3 Pro Image Preview
- `openrouter/openai/gpt-5-image` - GPT-5 Image
- `openrouter/openai/gpt-5-image-mini` - GPT-5 Image Mini

## Provider Details

### OpenAI
- **Best for**: High-quality artistic images, photorealistic content
- **Features**: HD quality, style control (vivid/natural), size options
- **Limits**: DALL-E 3 generates 1 image per request, DALL-E 2 up to 10
- **Pricing**: Pay per image, HD costs more

```bash
# DALL-E 3 with vivid style
llm-imager -m openai/dall-e-3 -p "cyberpunk cityscape at night" --quality hd --style vivid -o cyberpunk.png

# DALL-E 2 multiple images
llm-imager -m openai/dall-e-2 -p "minimalist logo design" -n 4 --size 512x512 -o logo.png
```

### Google Gemini
- **Best for**: Fast generation, good quality/speed balance
- **Features**: Aspect ratio control, image size options
- **Limits**: Rate limits apply based on API tier

```bash
# Gemini with aspect ratio
llm-imager -m google/gemini-2.5-flash-image -p "mountain landscape panorama" --aspect-ratio 16:9 -o panorama.png

# Square format
llm-imager -m google/gemini-2.5-flash-image -p "product photo of headphones" --aspect-ratio 1:1 -o product.png
```

### Stability AI
- **Best for**: Fine control over generation, negative prompts, artistic styles
- **Features**: Negative prompts, seed control, generation steps
- **Models**: Core (fast), Ultra (quality), SD3 (latest)

```bash
# Stable Image Ultra with negative prompt
llm-imager -m stability/stable-image-ultra -p "portrait of a woman, oil painting style" \
  --negative-prompt "blurry, low quality, distorted" -o portrait.png

# SD3 with seed for reproducibility
llm-imager -m stability/sd3-large -p "abstract geometric art" --seed 12345 -o abstract.png
```

### Replicate
- **Best for**: Access to open-source models, FLUX, SDXL
- **Features**: Many model variants, custom parameters
- **Note**: Generation may take longer due to cold starts

```bash
# FLUX 1.1 Pro
llm-imager -m replicate/flux-1.1-pro -p "hyperrealistic photo of a coffee cup" -o coffee.png

# FLUX Schnell (faster)
llm-imager -m replicate/flux-schnell -p "quick sketch of a cat" -o cat.png

# SDXL with aspect ratio
llm-imager -m replicate/sdxl -p "fantasy castle" --aspect-ratio 16:9 -o castle.png
```

### OpenRouter
- **Best for**: Single API key for multiple providers, fallback options
- **Features**: Access to various models through unified API
- **Note**: Pricing varies by underlying model

```bash
# GPT-5 Image via OpenRouter
llm-imager -m openrouter/openai/gpt-5-image -p "futuristic robot" -o robot.png

# Gemini via OpenRouter
llm-imager -m openrouter/google/gemini-2.5-flash-image -p "watercolor flowers" -o flowers.png
```

## Advanced Examples

### Batch Generation Script

```bash
#!/bin/bash
prompts=("sunset beach" "mountain forest" "city skyline")
for i in "${!prompts[@]}"; do
  llm-imager -m google/gemini-2.5-flash-image -p "${prompts[$i]}" -o "image_$i.png"
done
```

### Different Formats

```bash
# PNG (default, lossless)
llm-imager -p "logo design" -o logo.png

# JPEG (smaller file size)
llm-imager -p "photo landscape" -o photo.jpg

# WebP (modern format)
llm-imager -p "web banner" -o banner.webp
```

### Using Config File for Defaults

```yaml
# ~/.llm-imager.yaml
defaults:
  model: "openai/dall-e-3"
  quality: "hd"
  style: "vivid"

providers:
  openai:
    enabled: true
  google:
    enabled: true
```

Then simply:
```bash
llm-imager -p "your prompt" -o output.png
```

## Troubleshooting

### API Key Errors

```
Error: OpenAI API key is required
```
**Solution**: Set the environment variable:
```bash
export OPENAI_API_KEY="sk-your-key-here"
```

### Invalid Model ID

```
Error: model not found
```
**Solution**: Check available models:
```bash
llm-imager list models
```

### Rate Limits

```
Error: rate limit exceeded
```
**Solution**: Wait and retry, or use a different provider:
```bash
# Switch to another provider
llm-imager -m google/gemini-2.5-flash-image -p "your prompt" -o output.png
```

### Timeout Errors

```
Error: context deadline exceeded
```
**Solution**: Some models take longer. Increase timeout in config:
```yaml
providers:
  replicate:
    timeout: 300s
```

### Content Policy Violations

```
Error: content policy violation
```
**Solution**: Modify your prompt to comply with provider guidelines. Avoid explicit content, violence, or copyrighted characters.

## Contributing

Contributions are welcome! Here's how to get started:

### Development Setup

```bash
git clone https://github.com/piligrim/llm-imager.git
cd llm-imager
go mod download
go build ./...
```

### Running Tests

```bash
go test ./...
```

### Adding a New Provider

1. Create a new file in `internal/provider/`
2. Implement the `Provider` interface:
   - `Name() string`
   - `SupportedModels() []Model`
   - `ValidateRequest(*generator.Request) error`
   - `Generate(context.Context, *generator.Request) (*generator.Response, error)`
3. Register the provider in `internal/provider/registry.go`
4. Add tests in `internal/provider/<name>_test.go`

### Code Style

- Follow Go conventions (`gofmt`, `golint`)
- Add comments for exported functions
- Write tests for new functionality

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/new-provider`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit with clear messages
6. Push and create a Pull Request

## License

MIT
