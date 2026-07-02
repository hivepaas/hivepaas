package jsonutil

import (
	"encoding/json"
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		actual   string
		expected string
		want     bool
	}{
		{
			name:     "exact map match",
			actual:   `{"k1":"v1"}`,
			expected: `{"k1":"v1"}`,
			want:     true,
		},
		{
			name:     "contain map match",
			actual:   `{"k1":"v1","k2":"v2"}`,
			expected: `{"k1":"v1"}`,
			want:     true,
		},
		{
			name:     "missing key in map",
			actual:   `{"k1":"v1"}`,
			expected: `{"k2":"v2"}`,
			want:     false,
		},
		{
			name:     "nested exact match",
			actual:   `{"k1":{"k2":"v2"}}`,
			expected: `{"k1":{"k2":"v2"}}`,
			want:     true,
		},
		{
			name:     "nested contain match",
			actual:   `{"k1":{"k2":"v2","k3":"v3"},"k4":"v4"}`,
			expected: `{"k1":{"k2":"v2"}}`,
			want:     true,
		},
		{
			name:     "nested mismatch",
			actual:   `{"k1":{"k2":"v2"}}`,
			expected: `{"k1":{"k3":"v3"}}`,
			want:     false,
		},
		{
			name:     "array exact match",
			actual:   `[1,2,3]`,
			expected: `[1,2,3]`,
			want:     true,
		},
		{
			name:     "array contain match (false because array uses DeepEqual)",
			actual:   `[1,2,3]`,
			expected: `[1,2]`,
			want:     false,
		},
		{
			name:     "type mismatch",
			actual:   `{"k1":"v1"}`,
			expected: `["k1"]`,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual, expected any
			if err := json.Unmarshal([]byte(tt.actual), &actual); err != nil {
				t.Fatalf("failed to unmarshal actual: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.expected), &expected); err != nil {
				t.Fatalf("failed to unmarshal expected: %v", err)
			}

			if got := Contains(actual, expected); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
