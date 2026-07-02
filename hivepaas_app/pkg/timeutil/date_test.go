package timeutil

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDate(t *testing.T) {
	tm := time.Date(2023, 10, 27, 12, 30, 45, 0, time.FixedZone("test", 7*3600))
	date := NewDate(tm)
	expected := time.Date(2023, 10, 27, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, date.ToTime())
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		input    string
		expected Date
		wantErr  bool
	}{
		{"2023-10-27", NewDate(time.Date(2023, 10, 27, 0, 0, 0, 0, time.UTC)), false},
		{"2023-1-2", NewDate(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)), false},
		{"invalid", Date{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.expected.Equal(got))
			}
		})
	}
}

func TestDate_Basic(t *testing.T) {
	d1 := NewDate(time.Date(2023, 10, 27, 0, 0, 0, 0, time.UTC))
	d2 := NewDate(time.Date(2023, 10, 28, 0, 0, 0, 0, time.UTC))
	dZero := Date{}

	assert.True(t, d1.Equal(d1)) //nolint:gocritic
	assert.False(t, d1.Equal(d2))
	assert.True(t, d1.Before(d2))
	assert.False(t, d2.Before(d1))
	assert.True(t, d2.After(d1))
	assert.False(t, d1.After(d2))
	assert.True(t, dZero.IsZero())
	assert.False(t, d1.IsZero())

	d3 := d1.AddDate(0, 0, 1)
	assert.True(t, d2.Equal(d3))

	assert.Equal(t, 24*time.Hour, d2.Sub(d1))
}

func TestDate_String(t *testing.T) {
	date := NewDate(time.Date(2023, 10, 27, 0, 0, 0, 0, time.UTC))
	assert.Equal(t, "2023-10-27", date.String())
}

func TestDate_JSON(t *testing.T) {
	date := NewDate(time.Date(2023, 10, 27, 0, 0, 0, 0, time.UTC))

	// Marshal
	data, err := json.Marshal(date)
	assert.NoError(t, err)
	assert.Equal(t, `"2023-10-27"`, string(data))

	// Unmarshal
	var d Date
	err = json.Unmarshal(data, &d)
	assert.NoError(t, err)
	assert.True(t, date.Equal(d))

	// Unmarshal null
	err = json.Unmarshal([]byte("null"), &d)
	assert.NoError(t, err)
	assert.True(t, d.IsZero())
}

func TestDate_SQL(t *testing.T) {
	date := NewDate(time.Date(2023, 10, 27, 0, 0, 0, 0, time.UTC))

	// Value
	val, err := date.Value()
	assert.NoError(t, err)
	assert.Equal(t, date.ToTime(), val)

	// Value zero
	valZero, err := Date{}.Value()
	assert.NoError(t, err)
	assert.Nil(t, valZero)

	// Scan time.Time
	var d Date
	tm := time.Date(2023, 10, 27, 0, 0, 0, 0, time.UTC)
	err = d.Scan(tm)
	assert.NoError(t, err)
	assert.True(t, d.Equal(date))

	// Scan nil
	err = d.Scan(nil)
	assert.NoError(t, err)
	assert.True(t, d.IsZero())

	// Scan error
	err = d.Scan("invalid")
	assert.Error(t, err)
}
