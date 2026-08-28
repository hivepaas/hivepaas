package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidPhotoContent(t *testing.T) {
	// Valid PNG header (8 bytes) + dummy bytes
	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D}
	// Valid JPEG header
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}
	// Valid WebP header: RIFF + 4 bytes size + WEBP
	webpData := []byte{'R', 'I', 'F', 'F', 0x00, 0x00, 0x00, 0x00, 'W', 'E', 'B', 'P', 'V', 'P', '8'}
	// Invalid / Text / HTML
	htmlData := []byte("<!DOCTYPE html><html><body><script>alert(1)</script></body></html>")
	// Empty or too short
	shortData := []byte{0x89, 0x50}

	tests := []struct {
		name     string
		data     []byte
		ext      string
		expected bool
	}{
		{"valid png with .png ext", pngData, ".png", true},
		{"valid png with .PNG ext", pngData, ".PNG", true},
		{"valid png with .jpg ext (mismatch)", pngData, ".jpg", false},
		{"valid jpeg with .jpg ext", jpegData, ".jpg", true},
		{"valid jpeg with .jpeg ext", jpegData, ".jpeg", true},
		{"valid jpeg with .png ext (mismatch)", jpegData, ".png", false},
		{"valid webp with .webp ext", webpData, ".webp", true},
		{"html fake png", htmlData, ".png", false},
		{"html fake jpg", htmlData, ".jpg", false},
		{"too short data", shortData, ".png", false},
		{"empty data", nil, ".png", false},
		{"unsupported ext", pngData, ".gif", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPhotoContent(tt.data, tt.ext)
			assert.Equal(t, tt.expected, result)
		})
	}
}
