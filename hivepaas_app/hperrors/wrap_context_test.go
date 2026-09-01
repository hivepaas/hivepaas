package hperrors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Wrapping an HPError directly must hand back the very same object, so the long chain of
// `return hperrors.Wrap(err)` calls that propagate an error stays free of allocations and keeps
// anything the original call site attached to it.
func TestWrap_DirectHPErrorIsReturnedUnchanged(t *testing.T) {
	original := Wrap(ErrNotFound).WithParam("Name", "repo")

	assert.Same(t, original, Wrap(original))
	assert.Same(t, original, Wrap(Wrap(Wrap(original))))
}

// Adding context around an HPError used to be silently dropped: Wrap found the inner HPError and
// returned it, so the caller only ever saw the innermost message.
func TestWrap_KeepsContextAddedAroundHPError(t *testing.T) {
	inner := Wrap(errors.New("exit status 1"))
	outer := Wrap(fmt.Errorf("kopia repository create failed: %s (err: %w)",
		"found existing data in storage location", inner))

	assert.Contains(t, outer.Error(), "kopia repository create failed")
	assert.Contains(t, outer.Error(), "found existing data in storage location")
	assert.Contains(t, outer.Error(), "exit status 1")
}

// The identity of the error must survive the added context, otherwise a wrapped ErrNotFound would
// start rendering as a 500.
func TestWrap_ContextKeepsCodeAndStatus(t *testing.T) {
	inner := Wrap(ErrSettingNotFound).WithParam("Name", "backup-repo")
	outer := Wrap(fmt.Errorf("loading backup repo storage: %w", inner))

	assert.Equal(t, http.StatusNotFound, outer.StatusCode())
	assert.True(t, errors.Is(outer, ErrNotFound))
	assert.True(t, errors.Is(outer, ErrSettingNotFound))
	assert.Equal(t, inner.StatusCode(), outer.StatusCode())
}

// The copy must not share its param maps with the error it was built from.
func TestWrap_ContextDoesNotShareParamsWithInner(t *testing.T) {
	inner := Wrap(ErrNotFound).WithParam("Name", "repo")
	outer := Wrap(fmt.Errorf("while importing: %w", inner)).WithParam("Name", "snapshot")

	innerMsg, _ := inner.Message("en")
	outerMsg, _ := outer.Message("en")
	assert.NotEqual(t, innerMsg, outerMsg)
	assert.Contains(t, outerMsg, "snapshot")
	assert.Contains(t, innerMsg, "repo")
}

// hpError.Unwrap returns its inner error, so pointing an hpError at a chain that leads back to
// itself would make every chain walker spin forever. Nothing here may hang or overflow.
func TestWrap_ContextDoesNotCreateUnwrapCycle(t *testing.T) {
	err := Wrap(ErrNotFound)
	for range 20 {
		err = Wrap(fmt.Errorf("layer: %w", err))
	}

	assert.Equal(t, http.StatusNotFound, err.StatusCode())
	assert.True(t, errors.Is(err, ErrNotFound))
	assert.Contains(t, err.Error(), "layer:")
	assert.NotEmpty(t, err.Build("en").StackTrace)
}

// The stack trace should still point at where the error first arose, not where context was added.
func TestWrap_ContextKeepsOriginalStackTrace(t *testing.T) {
	inner := originatingCall()
	outer := Wrap(fmt.Errorf("added context: %w", inner))

	assert.Contains(t, outer.Build("en").StackTrace, "originatingCall")
}

func originatingCall() HPError {
	return Wrap(errors.New("something broke"))
}

// Extra detail attached to the inner error must survive too.
func TestWrap_ContextKeepsExtraDetail(t *testing.T) {
	inner := Wrap(ErrNotFound).WithExtraDetail("stderr: no such file")
	outer := Wrap(fmt.Errorf("listing snapshots: %w", inner))

	assert.Contains(t, outer.Build("en").Detail, "stderr: no such file")
}
