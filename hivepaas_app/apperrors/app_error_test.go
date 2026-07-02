package apperrors

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsValidMessageID(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"ERR_BAD_REQUEST", true},
		{"ERR_NOT_FOUND", true},
		{"ERR_123_ABC", true},
		{"ERR_", true},
		{"ERR_bad", false},          // lowercase not allowed
		{"ERR-BAD", false},          // dash not allowed
		{"NOT_ERR_", false},         // doesn't start with ERR_
		{"ERR_BAD!", false},         // special character not allowed
		{"ERR_BAD_REQUEST ", false}, // trailing space not allowed
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			actual := isValidMessageID(tc.input)
			if actual != tc.expected {
				t.Errorf("isValidMessageID(%q) = %v; want %v", tc.input, actual, tc.expected)
			}
		})
	}
}

func TestGetMessageID(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "standard error with message id",
			err:      errors.New("ERR_BAD_REQUEST"),
			expected: "ERR_BAD_REQUEST",
		},
		{
			name:     "standard error without message id",
			err:      errors.New("something went wrong"),
			expected: "",
		},
		{
			name:     "wrapped error with message id",
			err:      fmt.Errorf("wrapped error: %w", errors.New("ERR_BAD_REQUEST")),
			expected: "ERR_BAD_REQUEST",
		},
		{
			name:     "wrapped error without message id",
			err:      fmt.Errorf("wrapped error: %w", errors.New("something went wrong")),
			expected: "",
		},
		{
			name:     "double wrapped error with message id",
			err:      fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", errors.New("ERR_BAD_REQUEST"))),
			expected: "ERR_BAD_REQUEST",
		},
		{
			name:     "errors.Join containing message id",
			err:      errors.Join(errors.New("ordinary error"), errors.New("ERR_NOT_FOUND")),
			expected: "ERR_NOT_FOUND",
		},
		{
			name:     "errors.Join containing wrapped message id",
			err:      errors.Join(errors.New("ordinary error"), fmt.Errorf("wrap: %w", errors.New("ERR_NOT_FOUND"))),
			expected: "wrap: ERR_NOT_FOUND",
		},
		{
			name:     "errors.Join without message id",
			err:      errors.Join(errors.New("error A"), errors.New("error B")),
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := getMessageID(tc.err)
			if actual != tc.expected {
				t.Errorf("getMessageID(%v) = %q; want %q", tc.err, actual, tc.expected)
			}
		})
	}
}

func TestGetBaseError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected error
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: nil,
		},
		{
			name:     "direct mapped error",
			err:      ErrNotFound,
			expected: ErrNotFound,
		},
		{
			name:     "unmapped error",
			err:      errors.New("ERR_UNMAPPED_ERROR"),
			expected: nil,
		},
		{
			name:     "wrapped mapped error",
			err:      fmt.Errorf("wrap: %w", ErrNotFound),
			expected: ErrNotFound,
		},
		{
			name:     "double wrapped mapped error",
			err:      fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", ErrBadRequest)),
			expected: ErrBadRequest,
		},
		{
			name:     "errors.Join with mapped error first",
			err:      errors.Join(ErrNotFound, errors.New("unmapped")),
			expected: ErrNotFound,
		},
		{
			name:     "errors.Join with mapped error last",
			err:      errors.Join(errors.New("unmapped"), ErrBadRequest),
			expected: ErrBadRequest,
		},
		{
			name:     "errors.Join without mapped error",
			err:      errors.Join(errors.New("unmapped A"), errors.New("unmapped B")),
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := getBaseError(tc.err)
			if !errors.Is(actual, tc.expected) {
				t.Errorf("getBaseError(%v) = %v; want %v", tc.err, actual, tc.expected)
			}
		})
	}
}
