package slugify

import (
	"strings"

	"github.com/mozillazg/go-slugify"
)

var (
	defaultReplacements = []string{"-", "_"}
)

func Slugify(s string) string {
	return slugify.Slugify(s)
}

func SlugifyEx(s string, replacements []string, limit int) string {
	slug := slugify.Slugify(s)
	if len(replacements) > 0 {
		slug = strings.NewReplacer(replacements...).Replace(slug)
	}
	if limit < 0 || len(slug) <= limit {
		return slug
	}
	return slug[:limit]
}

func SlugifyAsKey(s string) string {
	return SlugifyEx(s, defaultReplacements, -1)
}
