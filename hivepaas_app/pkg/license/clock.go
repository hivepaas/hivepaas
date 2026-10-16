package license

import (
	"time"
)

// clockSlack is how far before the latest time seen the clock may read and still
// be believed. NTP and people correct clocks by seconds or minutes; a day is past
// any honest correction.
const clockSlack = day

// EffectiveNow is the time license checks go by: the clock, unless it reads more
// than a day before latestSeen, the latest time they have seen, in which case
// latestSeen. An offline installation can set its clock back to keep an expired
// key in force and nothing can stop that; this keeps it from being free.
func EffectiveNow(clock, latestSeen time.Time) time.Time {
	if clock.Before(latestSeen.Add(-clockSlack)) {
		return latestSeen
	}
	return clock
}
