package nanoid

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestNewStandard16_Length(t *testing.T) {
	// Ensure the generated ID is 16 characters long (bytes) and valid UTF-8
	id := NewStandard16()
	assert.NotEmpty(t, id, "generated nanoid should not be empty")
	// nanoid generates ASCII characters, so byte length equals rune count
	assert.Equal(t, 16, len(id), "generated nanoid should be 16 bytes long")
	// Verify valid UTF-8 (should be true for ASCII)
	assert.True(t, utf8.ValidString(id), "generated nanoid should be valid UTF-8")
}

func TestGeneratorStandard16_Unique(t *testing.T) {
	// Call the underlying generator directly to ensure it returns distinct values
	id1 := generatorStandard16()
	id2 := generatorStandard16()
	assert.NotEqual(t, id1, id2, "consecutive nanoid values should differ")
}
