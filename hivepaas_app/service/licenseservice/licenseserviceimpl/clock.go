package licenseserviceimpl

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
)

// seenRecordInterval is how far the clock moves on before the latest time seen
// is written again. A clock is believed down to a day before that time, so an
// hour behind costs nothing, and it keeps a write off nearly every read.
const seenRecordInterval = time.Hour

// clockState is the latest time license checks have seen, as this process last
// read or wrote it in system_statuses.
type clockState struct {
	loaded     bool
	latestSeen time.Time
}

// now is the time a key is evaluated at: the clock, or the latest time seen when
// the clock reads more than a day before it (license.EffectiveNow).
//
// The latest time seen is read from system_statuses once, and written there when
// the clock has moved on by an hour since. The write never lowers it and returns
// what is stored, so the app and the worker each keep up with the later of the
// two. A clock set back is never written.
func (s *service) now(ctx context.Context) (time.Time, error) {
	clock := s.timeNow().UTC()

	s.clockMu.Lock()
	defer s.clockMu.Unlock()
	if !s.clock.loaded {
		status, err := s.systemStatusRepo.Get(ctx, s.db)
		if err != nil {
			return time.Time{}, hperrors.Wrap(err)
		}
		s.clock = clockState{loaded: true, latestSeen: status.LicenseSeenAt.UTC()}
	}
	if clock.Sub(s.clock.latestSeen) >= seenRecordInterval {
		seen, err := s.systemStatusRepo.RecordLicenseSeenAt(ctx, s.db, clock)
		if err != nil {
			return time.Time{}, hperrors.Wrap(err)
		}
		s.clock.latestSeen = seen
	}
	return license.EffectiveNow(clock, s.clock.latestSeen), nil
}
