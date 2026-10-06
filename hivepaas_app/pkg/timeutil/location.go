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

// SameZone says two locations are one zone: the same name, or UTC both by
// different names - a server's Etc/UTC and HivePaaS's UTC.
func SameZone(a, b *time.Location) bool {
	return a.String() == b.String() || (alwaysUTC(a) && alwaysUTC(b))
}

// alwaysUTC says a location keeps UTC's clocks the year round - UTC, Etc/UTC,
// Etc/Universal - and is not a zone at UTC in winter only.
func alwaysUTC(loc *time.Location) bool {
	year := time.Now().Year()
	for _, month := range []time.Month{time.January, time.April, time.July, time.October} {
		if _, offset := time.Date(year, month, 1, 12, 0, 0, 0, loc).Zone(); offset != 0 {
			return false
		}
	}
	return true
}
