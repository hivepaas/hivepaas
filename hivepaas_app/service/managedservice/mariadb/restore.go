package mariadb

import (
	"regexp"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

// Where a restored MariaDB's replay of the binary logs stops, beside what every
// WAL-G engine's restore is told (walgkit, "What a restored app is restored
// with"): at a time (RFC 3339), at a point's position (FILE:OFFSET), or before
// it starts - the base backup as it is. None of them replays every log
// archived. walg-entrypoint's comment is the contract.
const (
	envRestoreStopTime     = walgkit.RestoreEnvPrefix + "STOP_TIME"
	envRestoreStopPosition = walgkit.RestoreEnvPrefix + "STOP_POSITION"
	envRestoreReplay       = walgkit.RestoreEnvPrefix + "REPLAY"
	replayNone             = "none"
)

// positionPattern is a point's position, as walg-mark prints it: the binary
// log and the offset of its end. It reaches the replay's arguments.
var positionPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+\.[0-9]{6,}:[0-9]+$`)

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
		if !positionPattern.MatchString(target.Position) || source.PointAt.IsZero() {
			return nil, targetInvalid("the point has no binary log position to replay to")
		}
		return [][2]string{{walgkit.EnvRestoreBefore, source.PointAt.UTC().Format(time.RFC3339)},
			{envRestoreStopPosition, target.Position}}, nil
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
