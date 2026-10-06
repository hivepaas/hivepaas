package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// UTC until set, and what was set after.
func TestLocation(t *testing.T) {
	t.Cleanup(func() { SetLocation(nil) })
	assert.Equal(t, time.UTC, Location())

	loc, err := time.LoadLocation("America/New_York")
	assert.NoError(t, err)
	SetLocation(loc)
	assert.Equal(t, loc, Location())
}
