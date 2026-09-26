package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

// MaxToolOutput is the most a tool answers. A model reads every byte it is
// given; an answer the size of a log file is one it cannot use.
const MaxToolOutput = 64 << 10

// shrinker is an answer that can be made smaller: a list that drops items from
// its end, and says how many it dropped. shrink reports false when there is
// nothing left to drop.
type shrinker interface {
	shrink() bool
}

// boundAnswer shrinks an answer until its JSON fits MaxToolOutput, or says it
// cannot be made to fit.
func boundAnswer(answer any) error {
	for {
		raw, err := json.Marshal(answer)
		if err != nil {
			return fmt.Errorf("mcp: encoding the answer: %w", err)
		}
		if len(raw) <= MaxToolOutput {
			return nil
		}
		s, ok := answer.(shrinker)
		if !ok || !s.shrink() {
			return &InputError{Message: fmt.Sprintf(
				"the answer is %d bytes, more than the %d a tool answers; narrow the request", len(raw), MaxToolOutput)}
		}
	}
}

// lastLines keeps the newest lines that fit in budget bytes, and says how many
// older ones it left out. Logs are read from the end: the last lines are the
// ones that say what happened.
func lastLines(lines []string, budget int) (kept []string, dropped int) {
	size := 0
	start := len(lines)
	for start > 0 {
		line := lines[start-1]
		if size+len(line)+1 > budget {
			break
		}
		size += len(line) + 1
		start--
	}
	if start == len(lines) && len(lines) > 0 {
		// One line longer than the whole budget: keep its end.
		line := lines[len(lines)-1]
		cut := len(line) - budget
		for cut < len(line) && !utf8.RuneStart(line[cut]) {
			cut++
		}
		return []string{"…" + line[cut:]}, len(lines) - 1
	}
	return lines[start:], start
}

// cutText keeps the first budget bytes of s, cut at a character boundary, with
// a note saying how much was left out and how to ask for it.
func cutText(s string, budget int, hint string) string {
	if len(s) <= budget {
		return s
	}
	cut := budget
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	var b strings.Builder
	b.WriteString(s[:cut])
	fmt.Fprintf(&b, "\n\ntruncated: %d bytes left out", len(s)-cut)
	if hint != "" {
		b.WriteString("; " + hint)
	}
	return b.String()
}

// shrinkList drops a tenth of a list from its end, at least one item, and
// counts what it dropped.
func shrinkList[T any](items *[]T, truncated *int) bool {
	if len(*items) == 0 {
		return false
	}
	n := max(1, len(*items)/10) //nolint:mnd
	*items = (*items)[:len(*items)-n]
	*truncated += n
	return true
}
