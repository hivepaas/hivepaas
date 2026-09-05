package hperrors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrLevel_ShouldRecord(t *testing.T) {
	assert.False(t, ErrLevelInfo.ShouldRecord())
	assert.True(t, ErrLevelWarn.ShouldRecord())
	assert.True(t, ErrLevelError.ShouldRecord())
}

// The level decides what gets stored, so each class of error has to land on the level
// it is meant to: anything the caller caused is INFO and dropped, anything the system
// broke is ERROR and kept.
func TestParseError_LevelByErrorClass(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ErrLevel
	}{
		// The caller's fault: told no, nothing to act on.
		{"validation", ErrValidation, ErrLevelInfo},
		{"bad request", ErrBadRequest, ErrLevelInfo},
		{"argument invalid", NewArgumentInvalid("pageLimit"), ErrLevelInfo},
		{"unauthorized", ErrUnauthorized, ErrLevelInfo},
		{"session expired", ErrSessionJWTExpired, ErrLevelInfo},
		{"forbidden", ErrForbidden, ErrLevelInfo},
		{"not found", ErrNotFound, ErrLevelInfo},
		{"already exists", ErrAlreadyExist, ErrLevelInfo},
		{"in use", ErrInUse, ErrLevelInfo},
		{"unsupported", ErrUnsupported, ErrLevelInfo},
		{"inactive", ErrInactive, ErrLevelInfo},
		{"password not strong enough", ErrPasswordNotMeetRequirements, ErrLevelInfo},

		// Ours: a bug or an outage.
		{"internal", ErrInternal, ErrLevelError},
		{"panic", ErrPanic, ErrLevelError},
		{"service unavailable", ErrServiceUnavailable, ErrLevelError},
		{"not implemented", ErrNotImplemented, ErrLevelError},

		// The specific 4xx errors that still mean something needs repairing.
		{"backup repo password out of sync", ErrBackupRepoPasswordOutOfSync, ErrLevelWarn},
		{"data version newer than system", ErrDataVerNewerThanSystemVer, ErrLevelWarn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, level := ParseError(tt.err, "")
			assert.Equal(t, tt.want, level)
		})
	}
}

// An error that is not one of ours at all is the one most worth keeping: it is
// something nobody anticipated. It must not be dropped for lack of a known status.
func TestParseError_UnknownErrorIsRecorded(t *testing.T) {
	_, level := ParseError(errors.New("some driver blew up"), "")
	assert.Equal(t, ErrLevelError, level)
	assert.True(t, level.ShouldRecord())
}

// The escalation has to survive wrapping, since that is how the error actually
// arrives from the layers below.
func TestParseError_WarnLevelSurvivesWrapping(t *testing.T) {
	err := Wrap(ErrBackupRepoPasswordOutOfSync).
		WithMsgLog("while changing the password of repo %s", "repo-1")

	_, level := ParseError(err, "")
	assert.Equal(t, ErrLevelWarn, level)
	assert.True(t, level.ShouldRecord())
}

// A sibling error built on the same base must stay at INFO: raising one specific
// error must not raise the whole status class with it.
func TestParseError_WarnLevelDoesNotLeakToSiblings(t *testing.T) {
	_, level := ParseError(ErrBackupRepoPasswordRequired, "")
	assert.Equal(t, ErrLevelInfo, level)
	assert.False(t, level.ShouldRecord())
}
