package hperrors

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// A client is sent the params an error names, and only those that are set; a
// translatable one stays in the message.
func TestBuildSendsOnlyTheExposedParams(t *testing.T) {
	err := Wrap(ErrPreconditionFailed).
		WithParam("Limit", "apps").
		WithParam("Max", 20).
		WithParam("Error", "dial tcp 10.0.0.5:5432: connection refused").
		WithTransParam("LimitName", "ERR_NOT_FOUND").
		ExposeParams("Limit", "Max", "LimitName", "Unset")

	assert.Equal(t, map[string]any{"Limit": "apps", "Max": 20}, err.Build("en").Params)
}

// Every other error carries none, and the field is left out of the answer.
func TestBuildSendsNoParamsUnlessExposed(t *testing.T) {
	for _, err := range []HPError{
		NewNotFound("App"),
		Wrap(ErrPreconditionFailed).WithParam("Error", "exit status 1"),
		Wrap(ErrNotFound).WithParam("Name", "app").ExposeParams("Unset"),
	} {
		info := err.Build("en")
		assert.Nil(t, info.Params)
		b, jsonErr := json.Marshal(info)
		assert.NoError(t, jsonErr)
		assert.NotContains(t, string(b), `"params"`)
	}
}

// Context added around an error keeps what it exposes. The two are separate
// errors from then on: what one goes on to expose, the other does not.
func TestWrapKeepsExposedParamsApart(t *testing.T) {
	inner := Wrap(ErrPreconditionFailed).WithParam("A", 1).WithParam("B", 2).WithParam("C", 3).
		ExposeParams("A", "Unset").ExposeParams("AlsoUnset")
	outer := Wrap(fmt.Errorf("adding context: %w", inner))

	_ = outer.ExposeParams("B")
	_ = inner.ExposeParams("C")

	assert.Equal(t, map[string]any{"A": 1, "C": 3}, inner.Build("en").Params)
	assert.Equal(t, map[string]any{"A": 1, "B": 2}, outer.Build("en").Params)
}
