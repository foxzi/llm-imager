package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/piligrim/llm-imager/internal/generator"
)

// Writer handles saving images to disk
type Writer struct {
	defaultFormat string
}

// NewWriter creates a new output writer
func NewWriter(defaultFormat string) *Writer {
	if defaultFormat == "" {
		defaultFormat = "png"
	}
	return &Writer{
		defaultFormat: defaultFormat,
	}
}

// Write saves images to the specified path
// Returns the list of saved file paths
func (w *Writer) Write(images []generator.Image, outputPath string) ([]string, error) {
	if len(images) == 0 {
		return nil, fmt.Errorf("no images to save")
	}

	// Ensure parent directory exists
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	savedPaths := make([]string, 0, len(images))

	for i, img := range images {
		path := w.generatePath(outputPath, i, len(images), img.Format)

		if err := os.WriteFile(path, img.Data, 0644); err != nil {
			return nil, fmt.Errorf("failed to write image %s: %w", path, err)
		}

		savedPaths = append(savedPaths, path)
	}

	return savedPaths, nil
}

// generatePath generates the output path for an image
func (w *Writer) generatePath(basePath string, index, total int, format string) string {
	if format == "" {
		format = w.defaultFormat
	}

	ext := filepath.Ext(basePath)
	base := strings.TrimSuffix(basePath, ext)

	// Use the format from the image if no extension provided
	if ext == "" {
		ext = "." + format
	}

	// If multiple images, add index
	if total > 1 {
		return fmt.Sprintf("%s_%d%s", base, index+1, ext)
	}

	return base + ext
}
