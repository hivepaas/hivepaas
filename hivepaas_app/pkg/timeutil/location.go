package timeutil

import (
	"sync/atomic"
	"time"
)

// location is the installation's timezone; nil until set, which reads as UTC.
var location atomic.Pointer[time.Location]

// Location is the installation's timezone: what a schedule's hours are read in
// - a cron expression's, a system job's time of day. Times are kept in UTC
// whatever it is. UTC until the config sets it.
func Location() *time.Location {
	if loc := location.Load(); loc != nil {
		return loc
	}
	return time.UTC
}

// SetLocation sets the installation's timezone. The config does, once loaded.
func SetLocation(loc *time.Location) {
	location.Store(loc)
}
