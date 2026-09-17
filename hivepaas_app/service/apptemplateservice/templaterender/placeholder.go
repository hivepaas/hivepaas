package templaterender

import (
	"fmt"
	"regexp"
	"strings"
)

// placeholderPattern matches ${{ ref }}, and $${{ ref }} as its escape. ${VAR}
// is not a placeholder: it is HivePaaS's own environment variable reference,
// and it must reach the app untouched.
var placeholderPattern = regexp.MustCompile(`\$?\$\{\{\s*([A-Za-z0-9_.]+)\s*\}\}`)

// Resolver answers one placeholder. keep asks for the placeholder to stay in the
// output as written - how a base render leaves secrets out.
type Resolver func(ref string) (value any, keep bool, err error)

// Substitute replaces placeholders in every string of a tree, returning a new
// tree. A string that is exactly one placeholder takes the resolved value's own
// type, so a size stays a size and a number stays a number; a placeholder inside
// a longer string is formatted into it.
func Substitute(node any, resolve Resolver) (any, error) {
	switch v := node.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, value := range v {
			substituted, err := Substitute(value, resolve)
			if err != nil {
				return nil, err
			}
			out[key] = substituted
		}
		return out, nil
	case []any:
		out := make([]any, len(v))
		for i, value := range v {
			substituted, err := Substitute(value, resolve)
			if err != nil {
				return nil, err
			}
			out[i] = substituted
		}
		return out, nil
	case string:
		return substituteString(v, resolve)
	default:
		return v, nil
	}
}

func substituteString(text string, resolve Resolver) (any, error) {
	matches := placeholderPattern.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return text, nil
	}

	whole := matches[0]
	if len(matches) == 1 && whole[0] == 0 && whole[1] == len(text) && !isEscaped(text, whole) {
		ref := text[whole[2]:whole[3]]
		value, keep, err := resolve(ref)
		if err != nil {
			return nil, err
		}
		if keep {
			return canonicalPlaceholder(ref), nil
		}
		return value, nil
	}

	var out strings.Builder
	last := 0
	for _, match := range matches {
		out.WriteString(text[last:match[0]])
		last = match[1]
		ref := text[match[2]:match[3]]
		if isEscaped(text, match) {
			out.WriteString(text[match[0]+1 : match[1]])
			continue
		}
		value, keep, err := resolve(ref)
		if err != nil {
			return nil, err
		}
		if keep {
			out.WriteString(canonicalPlaceholder(ref))
			continue
		}
		_, _ = fmt.Fprint(&out, value)
	}
	out.WriteString(text[last:])
	return out.String(), nil
}

func isEscaped(text string, match []int) bool {
	return strings.HasPrefix(text[match[0]:], "$$")
}

func canonicalPlaceholder(ref string) string {
	return "${{ " + ref + " }}"
}
