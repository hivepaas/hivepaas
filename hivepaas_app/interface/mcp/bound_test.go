package mcp

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// An answer too large is refused, in words a model can act on.
func TestAnswersAreBounded(t *testing.T) {
	assert.NoError(t, boundAnswer(map[string]string{"log": "small"}))
	big := map[string]string{"log": strings.Repeat("y", MaxToolOutput)}
	var inputErr *InputError
	if assert.ErrorAs(t, boundAnswer(big), &inputErr) {
		assert.Contains(t, inputErr.Message, "pageLimit")
	}
}

func TestLogsKeepTheirNewestLines(t *testing.T) {
	var lines []string
	for range 1000 {
		lines = append(lines, strings.Repeat("l", 99))
	}
	lines = append(lines, "the last line")
	kept, dropped := lastLines(lines, 10_000)
	assert.Equal(t, "the last line", kept[len(kept)-1])
	assert.Equal(t, len(lines), len(kept)+dropped)
	assert.LessOrEqual(t, len(strings.Join(kept, "\n")), 10_000)

	// One line longer than the budget keeps its end, on a character boundary.
	kept, dropped = lastLines([]string{"a", strings.Repeat("é", 20)}, 11)
	assert.Equal(t, 1, dropped)
	assert.True(t, utf8.ValidString(kept[0]))
	assert.True(t, strings.HasSuffix(kept[0], "é"))
}
