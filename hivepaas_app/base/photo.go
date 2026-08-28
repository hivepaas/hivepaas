package base

import (
	"net/http"
	"strings"
)

const (
	PhotoMaxSize          = 300 * 1024 // 300KB
	PhotoDataBase64MaxLen = 450 * 1024 // 450KB Base64 string max
)

var (
	AllPhotoFileExts = []string{".png", ".jpg", ".jpeg", ".webp"}
)

// IsValidPhotoContent checks if the binary data has valid magic bytes matching the file extension.
func IsValidPhotoContent(data []byte, fileExt string) bool {
	if len(data) < 12 { //nolint:mnd
		return false
	}
	detectedType := http.DetectContentType(data)
	switch strings.ToLower(fileExt) {
	case ".png":
		return detectedType == "image/png"
	case ".jpg", ".jpeg":
		return detectedType == "image/jpeg"
	case ".webp":
		return detectedType == "image/webp"
	default:
		return false
	}
}
