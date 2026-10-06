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

// The same zone by name, or two that are UTC by different names - a server's
// Etc/UTC and HivePaaS's UTC. A zone at UTC only part of the year is not.
func TestSameZone(t *testing.T) {
	zone := func(name string) *time.Location {
		loc, err := time.LoadLocation(name)
		assert.NoError(t, err, name)
		return loc
	}
	assert.True(t, SameZone(time.UTC, zone("Etc/UTC")))
	assert.True(t, SameZone(zone("Etc/UTC"), zone("Etc/Universal")))
	assert.True(t, SameZone(zone("America/New_York"), zone("America/New_York")))
	assert.False(t, SameZone(time.UTC, zone("Europe/London")), "UTC in winter only")
	assert.False(t, SameZone(zone("America/New_York"), zone("America/Toronto")), "by name, not by clocks")
}
