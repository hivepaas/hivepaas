package slugify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlugifyBasic(t *testing.T) {
	// Basic conversion: spaces to hyphens, lowercasing, remove punctuation
	result := Slugify("Hello World!")
	assert.Equal(t, "hello-world", result)
}

func TestSlugifyExReplacementsAndLimit(t *testing.T) {
	input := "Hello World!"
	// No replacements, limit shorter than result length
	assert.Equal(t, "hello-wor", SlugifyEx(input, nil, 9))

	// Replace hyphen with underscore, no limit (large)
	assert.Equal(t, "hello_world", SlugifyEx(input, []string{"-", "_"}, 20))

	// Replace and apply limit that truncates the result
	assert.Equal(t, "hello_wor", SlugifyEx(input, []string{"-", "_"}, 9))
}

func TestSlugifyExEmptyInputs(t *testing.T) {
	// Empty string should stay empty
	assert.Empty(t, Slugify(""))
	assert.Empty(t, SlugifyEx("", nil, 10))
	// Nil replacements should not panic
	assert.Equal(t, "test", SlugifyEx("test", nil, 10))
}
