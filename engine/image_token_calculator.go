package engine

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"strings"

	// Import image decoders
	_ "image/jpeg"
	_ "image/png"
)

// ImageTokenCalculator provides more accurate token calculation for images
type ImageTokenCalculator struct {
	// DetailMode can be "low" or "high" (auto for OpenAI)
	DetailMode string
}

// NewImageTokenCalculator creates a new image token calculator
func NewImageTokenCalculator() *ImageTokenCalculator {
	return &ImageTokenCalculator{
		DetailMode: "high", // Default to high detail for accuracy
	}
}

// CalculateImageTokens calculates tokens for an image based on OpenAI's actual pricing model
func (calc *ImageTokenCalculator) CalculateImageTokens(contentType, base64Data string) int {
	// For non-image files, use simple estimation
	if !strings.HasPrefix(contentType, "image/") {
		return calc.calculateNonImageTokens(contentType, len(base64Data))
	}

	// Try to get actual image dimensions
	width, height, err := calc.getImageDimensions(contentType, base64Data)
	if err != nil {
		// Fallback to size-based estimation
		return calc.calculateImageTokensBySize(len(base64Data))
	}

	// Use OpenAI's actual token calculation formula
	return calc.calculateOpenAIImageTokens(width, height)
}

// calculateOpenAIImageTokens implements OpenAI's vision model token calculation
func (calc *ImageTokenCalculator) calculateOpenAIImageTokens(width, height int) int {
	// OpenAI's token calculation for images:
	// 1. If low detail mode: always 85 tokens
	// 2. If high detail mode:
	//    - First, the image is scaled to fit within 2048x2048, maintaining aspect ratio
	//    - Then, the image is scaled such that the shortest side is 768px long
	//    - Finally, count how many 512px squares the image consists of
	//    - Each 512px square costs 170 tokens
	//    - Base cost is always 85 tokens

	if calc.DetailMode == "low" {
		return 85
	}

	// Step 1: Scale to fit within 2048x2048
	scaledWidth, scaledHeight := calc.scaleToFit(width, height, 2048, 2048)

	// Step 2: Scale so shortest side is 768px
	var finalWidth, finalHeight int
	if scaledWidth < scaledHeight {
		// Width is shorter
		scale := 768.0 / float64(scaledWidth)
		finalWidth = 768
		finalHeight = int(float64(scaledHeight) * scale)
	} else {
		// Height is shorter (or equal)
		scale := 768.0 / float64(scaledHeight)
		finalHeight = 768
		finalWidth = int(float64(scaledWidth) * scale)
	}

	// Step 3: Calculate number of 512px tiles
	tilesX := (finalWidth + 511) / 512  // Ceiling division
	tilesY := (finalHeight + 511) / 512 // Ceiling division
	totalTiles := tilesX * tilesY

	// Each tile costs 170 tokens, plus base 85 tokens
	return 85 + (totalTiles * 170)
}

// scaleToFit scales dimensions to fit within maxWidth and maxHeight while maintaining aspect ratio
func (calc *ImageTokenCalculator) scaleToFit(width, height, maxWidth, maxHeight int) (int, int) {
	if width <= maxWidth && height <= maxHeight {
		return width, height
	}

	scaleX := float64(maxWidth) / float64(width)
	scaleY := float64(maxHeight) / float64(height)
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}

	return int(float64(width) * scale), int(float64(height) * scale)
}

// getImageDimensions tries to extract image dimensions from base64 data
func (calc *ImageTokenCalculator) getImageDimensions(contentType, base64Data string) (width, height int, err error) {
	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Decode image to get dimensions
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image: %w", err)
	}

	return img.Width, img.Height, nil
}

// calculateImageTokensBySize provides fallback estimation based on data size
func (calc *ImageTokenCalculator) calculateImageTokensBySize(dataLength int) int {
	// Conservative estimation based on typical image compression ratios
	// Small images (<100KB base64): ~85-255 tokens
	// Medium images (100KB-1MB base64): ~255-765 tokens
	// Large images (>1MB base64): ~765-1700 tokens

	if dataLength < 100*1024 { // < 100KB
		return 170 // 85 base + 1 tile
	} else if dataLength < 1024*1024 { // < 1MB
		return 510 // 85 base + 2.5 tiles average
	} else if dataLength < 5*1024*1024 { // < 5MB
		return 935 // 85 base + 5 tiles average
	} else {
		return 1445 // 85 base + 8 tiles average
	}
}

// calculateNonImageTokens estimates tokens for non-image files
func (calc *ImageTokenCalculator) calculateNonImageTokens(contentType string, dataLength int) int {
	switch {
	case strings.HasPrefix(contentType, "application/pdf"):
		// PDF: rough estimation - 1 token per 4 characters of original content
		estimatedOriginalSize := (dataLength * 3) / 4 // Base64 decode
		return estimatedOriginalSize / 4
	case strings.HasPrefix(contentType, "text/"):
		// Text files: 1 token per ~4 characters
		estimatedOriginalSize := (dataLength * 3) / 4
		return estimatedOriginalSize / 4
	case strings.HasPrefix(contentType, "audio/"):
		// Audio: very rough estimation - OpenAI Whisper uses different pricing
		// For simplicity, estimate high token count to be conservative
		return dataLength / 100
	default:
		// Unknown file type: conservative high estimate
		return dataLength / 50
	}
}
