package strutil

import (
	"strings"

	"github.com/samber/lo"
	"github.com/tiendc/gofn"
)

func ToSnakeCase(str string) string {
	return lo.SnakeCase(str)
}

func ToPascalCase(str string) string {
	return lo.PascalCase(str)
}

func ToCamelCase(str string) string {
	return lo.CamelCase(str)
}

func CutShort(s string, maxLen int, padding string) string {
	if gofn.RuneLength(s) <= maxLen {
		return s
	}
	return string([]rune(s)[:maxLen]) + padding
}

func RemoveEmptyLines(str string, trimSpace bool) string {
	lines := strings.Split(str, "\n")
	cleanedLines := make([]string, 0, len(lines))

	for _, line := range lines {
		if trimSpace {
			// If after trimming spaces the line is empty, skip it
			if strings.TrimSpace(line) == "" {
				continue
			}
			cleanedLines = append(cleanedLines, line)
			continue
		}
		// When not trimming spaces, keep non-empty lines (including whitespace-only)
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}
	return strings.Join(cleanedLines, "\n")
}

func Cut(s, sep string) (string, string, bool) {
	if sep == "" {
		return s, "", false
	}
	return strings.Cut(s, sep)
}

func GetFirstLine(input string) string {
	if idx := strings.IndexByte(input, '\n'); idx >= 0 {
		return strings.TrimSuffix(input[:idx], "\r")
	}
	return input
}
