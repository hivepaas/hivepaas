package unit

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataSize_Values(t *testing.T) {
	size := MB
	assert.Equal(t, int64(1024*1024), size.Bytes())
	assert.Equal(t, 1024.0, size.KBytes())
	assert.Equal(t, 1.0, size.MBytes())
	assert.Equal(t, 1.0/1024.0, size.GBytes())

	sizeGB := GB + 512*MB
	assert.Equal(t, 1.5, sizeGB.GBytes())
}

func TestDataSize_String(t *testing.T) {
	tests := []struct {
		size     DataSize
		expected string
	}{
		{0, "0"},
		{B, "1b"},
		{KB, "1kb"},
		{MB, "1mb"},
		{GB, "1gb"},
		{TB, "1tb"},
		{PB, "1pb"},
		{EB, "1eb"},
		{KB + B, "1025b"},
		{MB + KB, "1025kb"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.size.String())
	}
}

func TestDataSize_HumanReadable(t *testing.T) {
	tests := []struct {
		size     DataSize
		expected string
	}{
		{500 * B, "500 B"},
		{1500 * B, "1.5 KB"},
		{1 * MB, "1.0 MB"},
		{1025 * KB, "1.0 MB"},
		{2 * GB, "2.0 GB"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.size.HumanReadable())
	}
}

func TestParseDataSize(t *testing.T) {
	tests := []struct {
		input    string
		expected DataSize
		wantErr  bool
	}{
		{"", 0, false},
		{"1b", B, false},
		{"1KB", KB, false},
		{"1 mb", MB, false},
		{"10gb", 10 * GB, false},
		{"1.5gb", 0, true}, // only integer values are supported by UnmarshalText
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDataSizeString(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestDataSize_JSON(t *testing.T) {
	size := 2 * GB

	// Marshal
	data, err := json.Marshal(size)
	assert.NoError(t, err)
	assert.Equal(t, `"2gb"`, string(data))

	// Unmarshal string
	var s DataSize
	err = json.Unmarshal([]byte(`"2gb"`), &s)
	assert.NoError(t, err)
	assert.Equal(t, size, s)

	// Unmarshal number
	err = json.Unmarshal([]byte(`2048`), &s)
	assert.NoError(t, err)
	assert.Equal(t, DataSize(2048), s)

	// Unmarshal null
	err = json.Unmarshal([]byte("null"), &s)
	assert.NoError(t, err)
	assert.Equal(t, DataSize(0), s)
}

func TestMustParseDataSize(t *testing.T) {
	assert.NotPanics(t, func() {
		MustParseDataSizeString("1gb")
	})
	assert.Panics(t, func() {
		MustParseDataSizeString("invalid")
	})
}

func TestDataSize_Truncate(t *testing.T) {
	tests := []struct {
		b        DataSize
		sz       DataSize
		expected DataSize
	}{
		{10 * B, 0, 10 * B},
		{10 * B, 3 * B, 9 * B},
		{10 * B, 11 * B, 0},
		{10 * B, -3 * B, 9 * B},
		{-10 * B, 3 * B, -9 * B},
		{1 * MB, 100 * KB, 1000 * KB}, // 1024KB truncated by 100KB is 1000KB
		{1 * GB, 512 * MB, 1 * GB},
		{1*GB + 100*MB, 512 * MB, 1 * GB},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.b.Truncate(tt.sz))
	}
}
