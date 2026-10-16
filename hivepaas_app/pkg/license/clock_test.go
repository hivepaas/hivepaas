package license

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEffectiveNow(t *testing.T) {
	latest := testNow
	for name, tc := range map[string]struct{ clock, latestSeen, want time.Time }{
		"nothing seen yet":                 {clock: latest, want: latest},
		"the clock ahead":                  {clock: latest.Add(time.Hour), latestSeen: latest, want: latest.Add(time.Hour)},
		"the clock set back a little":      {clock: latest.Add(-time.Hour), latestSeen: latest, want: latest.Add(-time.Hour)},
		"the clock set back exactly a day": {clock: latest.Add(-day), latestSeen: latest, want: latest.Add(-day)},
		"the clock set back over a day": {clock: latest.Add(-day - time.Second), latestSeen: latest,
			want: latest},
		"the clock set back a year": {clock: latest.AddDate(-1, 0, 0), latestSeen: latest, want: latest},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, EffectiveNow(tc.clock, tc.latestSeen))
		})
	}
}
