package fileutil

import (
	"mime"
	"strings"
)

func TypeByExtension(fileExt string) string {
	if fileExt == "" {
		return ""
	}
	if !strings.HasPrefix(fileExt, ".") {
		fileExt = "." + fileExt
	}
	return mime.TypeByExtension(strings.ToLower(fileExt))
}
