package executil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuoteHelpers(t *testing.T) {
	assert.True(t, IsSingleQuoted("'hello'"))
	assert.False(t, IsSingleQuoted("\"hello\""))
	assert.False(t, IsSingleQuoted("hello"))

	assert.True(t, IsDoubleQuoted("\"hello\""))
	assert.False(t, IsDoubleQuoted("'hello'"))
	assert.False(t, IsDoubleQuoted("hello"))

	assert.True(t, IsQuoted("'hello'"))
	assert.True(t, IsQuoted("\"hello\""))
	assert.False(t, IsQuoted("hello"))
}

func TestArgQuote(t *testing.T) {
	tests := []struct {
		arg      string
		expected string
	}{
		{"hello", "hello"},
		{"hello world", "'hello world'"},
		{"'already quoted'", "'already quoted'"},
		{"\"already double quoted\"", "\"already double quoted\""},
		{"it's", "it\\'s"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, ArgQuote(tt.arg))
	}
}

func TestCmdSplit(t *testing.T) {
	tests := []struct {
		cmd      string
		expected []string
		wantErr  bool
	}{
		{"echo hello", []string{"echo", "hello"}, false},
		{"echo 'hello world'", []string{"echo", "hello world"}, false},
		{"ls -l \"my file\"", []string{"ls", "-l", "my file"}, false},
		{"echo 'unclosed", nil, true},
	}

	for _, tt := range tests {
		got, err := CmdSplit(tt.cmd)
		if tt.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		}
	}
}
