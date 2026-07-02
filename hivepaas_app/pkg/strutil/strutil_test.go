package strutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCaseConversions(t *testing.T) {
	// Snake case
	assert.Equal(t, "hello_world", ToSnakeCase("HelloWorld"))
	// Pascal case
	assert.Equal(t, "HelloWorld", ToPascalCase("hello_world"))
	// Camel case
	assert.Equal(t, "helloWorld", ToCamelCase("hello_world"))
}

func TestCutShort(t *testing.T) {
	s := "abcdefghijklmnopqrstuvwxyz"
	// When length is less than max, returns original
	assert.Equal(t, s, CutShort(s, 100, "..."))

	// When longer, truncate and add padding
	truncated := CutShort(s, 5, "...")
	// Expect first 5 runes (a,b,c,d,e) plus padding
	assert.Equal(t, "abcde...", truncated)
}

func TestRemoveEmptyLines(t *testing.T) {
	input := "line1\n\n line2 \n   \nline3\n"
	// trimSpace true: remove blank and whitespace-only lines
	expectedTrim := "line1\n line2 \nline3"
	assert.Equal(t, expectedTrim, RemoveEmptyLines(input, true))

	// trimSpace false: keep lines that are not exactly empty (whitespace lines stay)
	expectedNoTrim := "line1\n line2 \n   \nline3"
	assert.Equal(t, expectedNoTrim, RemoveEmptyLines(input, false))
}

func TestGetFirstLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"single line no newline", "hello world", "hello world"},
		{"multiline unix", "first line\nsecond line\nthird", "first line"},
		{"multiline windows", "first line\r\nsecond line\r\nthird", "first line"},
		{"starts with newline", "\nsecond line", ""},
		{"only newline", "\n", ""},
		{"only carriage return and newline", "\r\n", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetFirstLine(tt.input))
		})
	}
}
