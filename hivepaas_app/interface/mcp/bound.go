package mcp

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

// MaxToolOutput is the most a tool answers. A model reads every byte it is
// given; an answer the size of a log file is one it cannot use.
const MaxToolOutput = 64 << 10

// boundAnswer says whether an answer's JSON fits MaxToolOutput. An answer is
// the endpoint's own and is not cut: one that does not fit is refused, in
// words a model can act on - a list asked again a smaller page at a time.
func boundAnswer(answer any) error {
	raw, err := json.Marshal(answer)
	if err != nil {
		return fmt.Errorf("mcp: encoding the answer: %w", err)
	}
	if len(raw) > MaxToolOutput {
		return &InputError{Message: fmt.Sprintf("the answer is %d bytes, more than the %d a tool answers; "+
			"narrow the request - a list with a smaller pageLimit, then pageOffset for the next page",
			len(raw), MaxToolOutput)}
	}
	return nil
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
