package postgres

import (
	"slices"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

const (
	// nightlyBackupArgs and nightlyRetainArgs are walg-nightly's two steps:
	// the base backup, which lists the backups kept as outBackups, and the
	// retention, which deletes what is before the backup it is given, then
	// verifies the chain.
	nightlyBackupArgs = "backup"
	nightlyRetainArgs = "retain"
	// maxRunsAhead bounds how far ahead a point's end is looked for: a point is
	// kept 90 days at most, and a nightly job runs once a day or more.
	maxRunsAhead = 400
)

func (e *Engine) NightlyBackupArgs() string {
	return nightlyBackupArgs
}

// Retention keeps the retainFull newest full backups, and every one a kept point
// needs (design §4, "Retention and kept points"). Without the backups listed, it
// deletes nothing.
func (e *Engine) Retention(req *managedmodel.RetentionReq) (*managedmodel.Retention, error) {
	pg := req.Managed.Database
	if pg == nil {
		return &managedmodel.Retention{Args: nightlyRetainArgs}, nil
	}
	boundary := RetentionBoundary(backupsOf(req.Outputs[outBackups]), pg.RetainFull, pg.RestorePoints, req.Now)
	if boundary == "" {
		return &managedmodel.Retention{Args: nightlyRetainArgs}, nil
	}
	return &managedmodel.Retention{Args: nightlyRetainArgs + " " + boundary, Boundary: boundary}, nil
}

// RetentionBoundary is the oldest full backup to keep: the older of the
// retainFull-th newest and the one before the oldest point still kept. It is
// empty when there are fewer than retainFull full backups, which keeps them all.
// A kept point no backup reaches keeps nothing: it cannot be restored as it is.
func RetentionBoundary(backups []managedmodel.BackupRef, retainFull int,
	points []*entity.ManagedRestorePoint, now time.Time) string {
	fulls := fullBackups(backups)
	if retainFull <= 0 || len(fulls) < retainFull {
		return ""
	}
	boundary := fulls[len(fulls)-retainFull]
	for _, point := range points {
		if !point.KeepUntil.After(now) {
			continue
		}
		if base := baseBefore(fulls, point.At); base != nil && base.At.Before(boundary.At) {
			boundary = *base
		}
	}
	return boundary.Name
}

// fullBackups are the backups that are not deltas, oldest first.
func fullBackups(backups []managedmodel.BackupRef) []managedmodel.BackupRef {
	fulls := slices.DeleteFunc(slices.Clone(backups), func(b managedmodel.BackupRef) bool {
		return strings.Contains(b.Name, "_D_")
	})
	slices.SortStableFunc(fulls, func(a, b managedmodel.BackupRef) int { return a.At.Compare(b.At) })
	return fulls
}

// baseBefore is the latest full backup that starts before at, nil for none.
func baseBefore(fulls []managedmodel.BackupRef, at time.Time) *managedmodel.BackupRef {
	var base *managedmodel.BackupRef
	for i := range fulls {
		if !fulls[i].At.After(at) {
			base = &fulls[i]
		}
	}
	return base
}

// pointsStatus is each point, and until when it can be restored: until the
// first nightly run after which the backup before it is neither among the
// retainFull newest nor before a point still kept then.
func pointsStatus(points []*entity.ManagedRestorePoint, backups []managedmodel.BackupRef, retainFull int,
	nightly *entity.SchedJobSchedule, now time.Time) []*managedmodel.RestorePointStatus {
	fulls := fullBackups(backups)
	var runs []time.Time
	if nightly != nil && len(points) > 0 {
		runs, _ = nightly.CalcNextRuns(now, maxRunsAhead)
	}
	out := make([]*managedmodel.RestorePointStatus, 0, len(points))
	for _, point := range points {
		status := &managedmodel.RestorePointStatus{Point: point, Kept: point.KeepUntil.After(now)}
		if base := baseBefore(fulls, point.At); base != nil {
			status.RestorableUntil = lastRestorable(base, fulls, points, retainFull, runs)
		}
		out = append(out, status)
	}
	slices.SortStableFunc(out, func(a, b *managedmodel.RestorePointStatus) int {
		return a.Point.At.Compare(b.Point.At)
	})
	return out
}

// lastRestorable is the nightly run that deletes base, nil when none of the runs
// ahead does.
func lastRestorable(base *managedmodel.BackupRef, fulls []managedmodel.BackupRef,
	points []*entity.ManagedRestorePoint, retainFull int, runs []time.Time) *time.Time {
	newer := 0
	for _, full := range fulls {
		if full.At.After(base.At) {
			newer++
		}
	}
	for i, run := range runs {
		// Each run takes a full backup before it deletes.
		if newer+i+1 < retainFull {
			continue
		}
		held := false
		for _, point := range points {
			if !point.KeepUntil.After(run) {
				continue
			}
			if other := baseBefore(fulls, point.At); other != nil && !other.At.After(base.At) {
				held = true
				break
			}
		}
		if !held {
			return &run
		}
	}
	return nil
}

// heldBack is what kept points make retention keep beyond the retainFull
// newest full backups, nil when nothing.
func heldBack(backups []managedmodel.BackupRef, retainFull int, points []*entity.ManagedRestorePoint,
	now time.Time) *managedmodel.HeldBack {
	fulls := fullBackups(backups)
	boundary := RetentionBoundary(backups, retainFull, points, now)
	if boundary == "" || boundary == fulls[len(fulls)-retainFull].Name {
		return nil
	}
	held := &managedmodel.HeldBack{}
	for _, full := range fulls[:len(fulls)-retainFull] {
		if full.Name == boundary || held.Backups > 0 {
			held.Backups++
			held.Bytes += full.Bytes
		}
	}
	return held
}

func (e *Engine) StaleDeployPoints(managed *entity.AppManagedSettings, outputs map[string]string) []string {
	if managed == nil || managed.Database == nil {
		return nil
	}
	// Without the list there is no telling: an empty one would make every point
	// stale.
	fulls := fullBackups(backupsOf(outputs[outBackups]))
	if len(fulls) == 0 {
		return nil
	}
	var stale []string
	for _, point := range managed.Database.RestorePoints {
		if point.Origin == entity.ManagedPointOriginDeploy && baseBefore(fulls, point.At) == nil {
			stale = append(stale, point.Name)
		}
	}
	return stale
}
