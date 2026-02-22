package provider

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strconv"
	"strings"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/piligrim/llm-imager/internal/generator"
)

// DryRun implements a mock provider for testing without API calls
type DryRun struct{}

// NewDryRun creates a new DryRun provider
func NewDryRun() *DryRun {
	return &DryRun{}
}

func (d *DryRun) Name() string {
	return "dryrun"
}

func (d *DryRun) SupportedModels() []Model {
	return []Model{
		{
			ID:       "dryrun/placeholder",
			Name:     "Placeholder Generator",
			Provider: "dryrun",
			Sizes:    []string{"256x256", "512x512", "1024x1024"},
			Features: []string{},
		},
	}
}

func (d *DryRun) ValidateRequest(req *generator.Request) error {
	return nil
}

func (d *DryRun) Generate(ctx context.Context, req *generator.Request) (*generator.Response, error) {
	start := time.Now()

	width, height := parseSize(req.Size)
	count := req.Count
	if count <= 0 {
		count = 1
	}

	images := make([]generator.Image, count)
	for i := 0; i < count; i++ {
		data, err := generatePlaceholder(width, height, req.Prompt)
		if err != nil {
			return nil, err
		}
		images[i] = generator.Image{
			Data:   data,
			Format: "png",
			Width:  width,
			Height: height,
			Index:  i,
		}
	}

	return &generator.Response{
		Images:        images,
		Model:         "dryrun/placeholder",
		Provider:      "dryrun",
		RevisedPrompt: req.Prompt,
		GeneratedAt:   time.Now(),
		Duration:      time.Since(start),
	}, nil
}

// parseSize parses size string like "1024x1024" into width and height
func parseSize(size string) (int, int) {
	if size == "" {
		return 512, 512
	}

	parts := strings.Split(strings.ToLower(size), "x")
	if len(parts) != 2 {
		return 512, 512
	}

	width, err1 := strconv.Atoi(parts[0])
	height, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || width <= 0 || height <= 0 {
		return 512, 512
	}

	return width, height
}

// generatePlaceholder creates a placeholder PNG image with prompt text
func generatePlaceholder(width, height int, prompt string) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with light gray background
	bgColor := color.RGBA{R: 240, G: 240, B: 240, A: 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Draw diagonal lines
	lineColor := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	for i := 0; i < width && i < height; i++ {
		img.Set(i, i, lineColor)
		img.Set(width-1-i, i, lineColor)
	}

	// Draw border
	borderColor := color.RGBA{R: 180, G: 180, B: 180, A: 255}
	for x := range width {
		img.Set(x, 0, borderColor)
		img.Set(x, height-1, borderColor)
	}
	for y := range height {
		img.Set(0, y, borderColor)
		img.Set(width-1, y, borderColor)
	}

	// Draw prompt text
	textColor := color.RGBA{R: 60, G: 60, B: 60, A: 255}
	drawText(img, prompt, width, height, textColor)

	// Draw "DRY-RUN" label at bottom
	labelColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	drawLabel(img, "DRY-RUN", width, height, labelColor)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// drawText draws wrapped text on the image
func drawText(img *image.RGBA, text string, width, height int, col color.Color) {
	face := basicfont.Face7x13
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
	}

	// Wrap text to fit width
	lines := wrapText(text, width, face)

	// Calculate vertical centering
	lineHeight := face.Metrics().Height.Ceil()
	totalHeight := lineHeight * len(lines)
	startY := (height - totalHeight) / 2

	// Draw each line centered
	for i, line := range lines {
		lineWidth := font.MeasureString(face, line).Ceil()
		x := (width - lineWidth) / 2
		y := startY + (i+1)*lineHeight

		d.Dot = fixed.Point26_6{
			X: fixed.I(x),
			Y: fixed.I(y),
		}
		d.DrawString(line)
	}
}

// wrapText wraps text to fit within maxWidth
func wrapText(text string, maxWidth int, face font.Face) []string {
	const padding = 20
	effectiveWidth := maxWidth - padding*2

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		lineWidth := font.MeasureString(face, testLine).Ceil()
		if lineWidth > effectiveWidth && currentLine != "" {
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			currentLine = testLine
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	// Limit number of lines
	maxLines := 15
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines[maxLines-1] += "..."
	}

	return lines
}

// drawLabel draws a label at the bottom of the image
func drawLabel(img *image.RGBA, label string, width, height int, col color.Color) {
	face := basicfont.Face7x13
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
	}

	labelWidth := font.MeasureString(face, label).Ceil()
	x := (width - labelWidth) / 2
	y := height - 10

	d.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}
	d.DrawString(label)
}
