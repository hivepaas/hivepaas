package hperrors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// What a task's output tells a person of an error: its code, then what it
// means - not the chain of codes Error() is.
func TestGetErrorDetailIsTheCodeAndWhatItMeans(t *testing.T) {
	err := NewMissing("Registry auth to pull image")

	assert.Equal(t, "ERR_PRECONDITION_FAILED\nERR_MISSING", err.Error())
	assert.Equal(t, "ERR_MISSING\nRegistry auth to pull image is missing", GetErrorDetail(err, ""))
}

// An error no code names is its message alone, not one under an empty line.
func TestGetErrorDetailOfAnErrorWithNoCodeIsItsMessage(t *testing.T) {
	assert.Equal(t, "exit status 2", GetErrorDetail(errors.New("exit status 2"), ""))
}
