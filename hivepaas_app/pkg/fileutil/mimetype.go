package fileutil

import (
	"mime"
	"strings"
)

func TypeByExtension(fileExt string) string {
	if !strings.HasPrefix(fileExt, ".") {
		fileExt = "." + fileExt
	}
	return mime.TypeByExtension(fileExt)
}
