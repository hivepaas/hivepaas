package mcp

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

type testList struct {
	Items     []string `json:"items"`
	Truncated int      `json:"truncated,omitempty"`
}

func (l *testList) shrink() bool {
	if len(l.Items) == 0 {
		return false
	}
	n := max(1, len(l.Items)/10)
	l.Items = l.Items[:len(l.Items)-n]
	l.Truncated += n
	return true
}

func TestAnswersAreBounded(t *testing.T) {
	list := &testList{}
	for range 5000 {
		list.Items = append(list.Items, strings.Repeat("x", 100))
	}
	assert.NoError(t, boundAnswer(list))
	assert.Less(t, len(list.Items), 5000)
	assert.Equal(t, 5000, len(list.Items)+list.Truncated, "what was dropped is counted")

	// An answer that cannot shrink says so, in words a model can act on.
	big := map[string]string{"log": strings.Repeat("y", MaxToolOutput)}
	var inputErr *InputError
	assert.ErrorAs(t, boundAnswer(big), &inputErr)
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

func TestCutTextNeverSplitsACharacter(t *testing.T) {
	cut := cutText(strings.Repeat("日本", 100), 101, "ask for less")
	assert.True(t, utf8.ValidString(cut))
	assert.Contains(t, cut, "truncated:")
	assert.Contains(t, cut, "ask for less")
	assert.Equal(t, "short", cutText("short", 100, ""))
}
