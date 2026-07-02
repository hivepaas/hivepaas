package redact

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	// Test if New correctly builds a Redactor
	r := New([]string{"secret", "password"})
	if r == nil {
		t.Fatal("Expected non-nil Redactor")
	}
	if r.replacer == nil {
		t.Fatal("Expected non-nil replacer inside Redactor")
	}
}

func TestRedactor_String(t *testing.T) {
	tests := []struct {
		name     string
		secrets  []string
		input    string
		expected string
	}{
		{
			name:     "Basic redaction",
			secrets:  []string{"mysecret", "password"},
			input:    "my password is password and secret is mysecret",
			expected: "my ******** is ******** and secret is ********",
		},
		{
			name:     "Empty secrets list",
			secrets:  []string{},
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "Overlapping secrets length sorting check",
			secrets:  []string{"secret", "secret_key"},
			input:    "the key is secret_key and word is secret",
			expected: "the key is ******** and word is ********",
		},
		{
			name:     "No matches",
			secrets:  []string{"secret"},
			input:    "hello world",
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.secrets)
			actual := r.String(tt.input)
			if actual != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

func TestRedactor_Slice(t *testing.T) {
	secrets := []string{"secret", "password"}
	r := New(secrets)

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Small slice (below concurrency threshold)",
			input:    []string{"my secret log", "public log", "password entry"},
			expected: []string{"my ******** log", "public log", "******** entry"},
		},
		{
			name:     "Empty slice",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := r.Slice(tt.input)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, actual)
			}
		})
	}
}

func TestRedactor_SliceParallel(t *testing.T) {
	// To test the parallel branch, we must provide a slice larger than concurrencyThreshold (500)
	size := concurrencyThreshold + 50
	input := make([]string, size)
	expected := make([]string, size)
	for i := 0; i < size; i++ {
		if i%2 == 0 {
			input[i] = fmt.Sprintf("log %d with secret", i)
			expected[i] = fmt.Sprintf("log %d with ********", i)
		} else {
			input[i] = fmt.Sprintf("log %d clean", i)
			expected[i] = fmt.Sprintf("log %d clean", i)
		}
	}

	r := New([]string{"secret"})
	actual := r.Slice(input)

	if len(actual) != size {
		t.Fatalf("Expected slice of size %d, got %d", size, len(actual))
	}

	for i := 0; i < size; i++ {
		if actual[i] != expected[i] {
			t.Errorf("At index %d: expected %q, got %q", i, expected[i], actual[i])
		}
	}
}
