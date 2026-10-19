package postgres

import (
	"regexp"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

// Where a restored Postgres recovers to, beside what every WAL-G engine's
// restore is told (walgkit, "What a restored app is restored with"):
// envRestoreRecovery is the recovery settings walg-entrypoint writes,
// space-separated name=value pairs; recovery_target_action=promote is always
// added.
const (
	envRestoreRecovery = walgkit.RestoreEnvPrefix + "RECOVERY"
	// recoveryTimeLayout is a time recovery_target_time takes: ISO 8601 with a
	// numeric offset. Postgres refuses the RFC 3339 Z, as it does any zone name.
	recoveryTimeLayout = "2006-01-02T15:04:05-07:00"
)

var (
	// lsnPattern and backupPattern are what a target may say, before it reaches
	// the server's configuration or WAL-G's arguments.
	lsnPattern    = regexp.MustCompile(`^[0-9A-Fa-f]{1,8}/[0-9A-Fa-f]{1,8}$`)
	backupPattern = regexp.MustCompile(`^base_[0-9A-Fa-f]{24}(_D_[0-9A-Fa-f]{24})?$`)
)

// targetEnv is a target as what to fetch and how to recover (design §4, "What
// to restore to").
func targetEnv(source *managedmodel.RestoreSource) ([][2]string, error) {
	target := source.Target
	switch target.Kind {
	case entity.ManagedRestoreToLatest:
		return [][2]string{{walgkit.EnvRestoreBackup, walgkit.RestoreLatest}}, nil
	case entity.ManagedRestoreToTime:
		if target.Time.IsZero() {
			return nil, targetInvalid("a time is required")
		}
		at := target.Time.UTC()
		return [][2]string{{walgkit.EnvRestoreBefore, at.Format(time.RFC3339)},
			{envRestoreRecovery, "recovery_target_time=" + at.Format(recoveryTimeLayout)}}, nil
	case entity.ManagedRestoreToPoint:
		if !lsnPattern.MatchString(target.Position) || source.PointAt.IsZero() {
			return nil, targetInvalid("the point has no LSN to recover to")
		}
		return [][2]string{
			{walgkit.EnvRestoreBefore, source.PointAt.UTC().Format(time.RFC3339)},
			{envRestoreRecovery, "recovery_target_lsn=" + target.Position + " recovery_target_inclusive=on"},
		}, nil
	case entity.ManagedRestoreToBackup:
		if !backupPattern.MatchString(target.Backup) {
			return nil, targetInvalid("not a base backup's name")
		}
		return [][2]string{{walgkit.EnvRestoreBackup, target.Backup},
			{envRestoreRecovery, "recovery_target=immediate"}}, nil
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
