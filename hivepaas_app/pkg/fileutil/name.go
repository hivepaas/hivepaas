package fileutil

import (
	"regexp"
	"strings"
)

var (
	invalidFilenameCharsRegex = regexp.MustCompile(`[\\/:*?"<>|\x00-\x1f]`)
)

func SanitizeFileName(name string) string {
	name = strings.ReplaceAll(name, "*", "wildcard")
	name = invalidFilenameCharsRegex.ReplaceAllString(name, "_")
	name = strings.Trim(name, " ._")
	return name
}
