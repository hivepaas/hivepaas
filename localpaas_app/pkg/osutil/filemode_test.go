package osutil

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFileMode(t *testing.T) {
	tests := []struct {
		input    string
		expected FileMode
		wantErr  bool
	}{
		{"0644", 0644, false},
		{"644", 0644, false},
		{"0755", 0755, false},
		{"755", 0755, false},
		{"0", 0, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseFileMode(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestFileMode_String(t *testing.T) {
	assert.Equal(t, "0644", FileMode(0644).String())
	assert.Equal(t, "0755", FileMode(0755).String())
	assert.Equal(t, "00", FileMode(0).String())
}

func TestFileMode_ToFileMode(t *testing.T) {
	assert.Equal(t, os.FileMode(0644), FileMode(0644).ToFileMode())
}

func TestFileMode_MarshalJSON(t *testing.T) {
	fm := FileMode(0644)
	data, err := json.Marshal(fm)
	assert.NoError(t, err)
	assert.Equal(t, `"0644"`, string(data))
}

func TestFileMode_UnmarshalJSON(t *testing.T) {
	t.Run("Valid octal string", func(t *testing.T) {
		var fm FileMode
		err := json.Unmarshal([]byte(`"0755"`), &fm)
		assert.NoError(t, err)
		assert.Equal(t, FileMode(0755), fm)
	})

	t.Run("Null value", func(t *testing.T) {
		var fm FileMode = 0644
		err := json.Unmarshal([]byte(`null`), &fm)
		assert.NoError(t, err)
		assert.Equal(t, FileMode(0), fm)
	})

	t.Run("Invalid string", func(t *testing.T) {
		var fm FileMode
		err := json.Unmarshal([]byte(`"invalid"`), &fm)
		assert.Error(t, err)
	})
}
