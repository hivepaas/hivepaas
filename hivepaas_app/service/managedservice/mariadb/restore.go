package mariadb

import (
	"regexp"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// Where a restored MariaDB's replay of the binary logs stops, beside what every
// WAL-G engine's restore is told (walgkit, "What a restored app is restored
// with"): at a time (RFC 3339), at a point's GTID (a GTID list, which it
// includes), at a point's FILE:OFFSET (one made before points kept their GTID),
// or before it starts - the base backup as it is. None of them replays every log
// archived. walg-entrypoint's comment is the contract.
const (
	envRestoreStopTime     = walgkit.RestoreEnvPrefix + "STOP_TIME"
	envRestoreStopGTID     = walgkit.RestoreEnvPrefix + "STOP_GTID"
	envRestoreStopPosition = walgkit.RestoreEnvPrefix + "STOP_POSITION"
	envRestoreReplay       = walgkit.RestoreEnvPrefix + "REPLAY"
	replayNone             = "none"
)

// A point's position, as walg-mark prints it. Both reach the replay's
// arguments.
var (
	// gtidPattern is a GTID list, one per replication domain: what a point's
	// position is.
	gtidPattern = regexp.MustCompile(`^[0-9]+-[0-9]+-[0-9]+(,[0-9]+-[0-9]+-[0-9]+)*$`)
	// positionPattern is the binary log and the offset of its end: a point's
	// position before points kept their GTID, and one made before any
	// transaction.
	positionPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+\.[0-9]{6,}:[0-9]+$`)
)

// gtidStopScript is the script that takes envRestoreStopGTID, and what it
// mentions when it does: a template's scripts made before they did would
// replay past the point, to the last log archived.
const gtidStopScript = "walg-entrypoint"

// checkTarget refuses a target the template's scripts would not stop at.
func checkTarget(doc *specmodel.AppDoc, source *managedmodel.RestoreSource) error {
	if source.Target.Kind == entity.ManagedRestoreToPoint && gtidPattern.MatchString(source.Target.Position) &&
		!walgkit.ScriptMentions(doc, gtidStopScript, envRestoreStopGTID) {
		return targetInvalid("the template's scripts cannot stop at a point's GTID: update the template, " +
			"or restore to the point's time")
	}
	return nil
}

// targetEnv is a target as what to fetch and where the replay stops.
func targetEnv(source *managedmodel.RestoreSource) ([][2]string, error) {
	target := source.Target
	switch target.Kind {
	case entity.ManagedRestoreToLatest:
		return [][2]string{{walgkit.EnvRestoreBackup, walgkit.RestoreLatest}}, nil
	case entity.ManagedRestoreToTime:
		if target.Time.IsZero() {
			return nil, targetInvalid("a time is required")
		}
		at := target.Time.UTC().Format(time.RFC3339)
		return [][2]string{{walgkit.EnvRestoreBefore, at}, {envRestoreStopTime, at}}, nil
	case entity.ManagedRestoreToPoint:
		stop := envRestoreStopGTID
		if !gtidPattern.MatchString(target.Position) {
			stop = envRestoreStopPosition
			if !positionPattern.MatchString(target.Position) {
				return nil, targetInvalid("the point has no binary log position to replay to")
			}
		}
		if source.PointAt.IsZero() {
			return nil, targetInvalid("the point has no binary log position to replay to")
		}
		return [][2]string{{walgkit.EnvRestoreBefore, source.PointAt.UTC().Format(time.RFC3339)},
			{stop, target.Position}}, nil
	case entity.ManagedRestoreToBackup:
		if !backupPattern.MatchString(target.Backup) {
			return nil, targetInvalid("not a base backup's name")
		}
		return [][2]string{{walgkit.EnvRestoreBackup, target.Backup}, {envRestoreReplay, replayNone}}, nil
	}
	return nil, targetInvalid("the target is latest, time, point or backup")
}

func targetInvalid(detail string) error {
	return hperrors.Wrap(hperrors.ErrAppManagedRestoreTargetInvalid).WithExtraDetail("%s", detail)
}

func (e *Engine) RecoveryCheck() managedmodel.RecoveryCheck {
	return walgkit.RecoveryCheck()
}

func (e *Engine) FirstBackupJob(managed *entity.AppManagedSettings) string {
	return walgkit.FirstBackupJob(managed)
}

func (e *Engine) FinishRestore(req *managedmodel.FinishRestoreReq) (*managedmodel.FinishRestoreResp, error) {
	resp, err := walgkit.FinishRestore(profile, req)
	return resp, hperrors.Wrap(err)
}
