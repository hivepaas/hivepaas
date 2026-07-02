package timeutil

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected Duration
		wantErr  bool
	}{
		{"1h30m", Duration(time.Hour + 30*time.Minute), false},
		{"1d", Duration(24 * time.Hour), false},
		{"1w", Duration(7 * 24 * time.Hour), false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDuration(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestDuration_String(t *testing.T) {
	tests := []struct {
		dur      Duration
		expected string
	}{
		{Duration(30 * time.Minute), "30m"},
		{Duration(25 * time.Hour), "1d1h"},
		{Duration(2 * 24 * time.Hour), "2d"},
		{Duration(-(25 * time.Hour)), "-1d1h"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.dur.String())
	}
}

func TestDuration_ToDuration(t *testing.T) {
	dur := Duration(time.Hour)
	assert.Equal(t, time.Hour, dur.ToDuration())
}

func TestDuration_JSON(t *testing.T) {
	dur := Duration(25 * time.Hour)

	// Marshal
	data, err := json.Marshal(dur)
	assert.NoError(t, err)
	assert.Equal(t, `"1d1h"`, string(data))

	// Unmarshal string
	var d Duration
	err = json.Unmarshal([]byte(`"1d1h"`), &d)
	assert.NoError(t, err)
	assert.Equal(t, dur, d)

	// Unmarshal number (nanoseconds)
	err = json.Unmarshal([]byte(`3600000000000`), &d)
	assert.NoError(t, err)
	assert.Equal(t, Duration(time.Hour), d)

	// Unmarshal null
	err = json.Unmarshal([]byte("null"), &d)
	assert.NoError(t, err)
	assert.Equal(t, Duration(0), d)
}
