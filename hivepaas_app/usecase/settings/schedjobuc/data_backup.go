package schedjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// checkDataBackup refuses a data backup that is not its app's: another app's
// command, or a volume the app does not mount as its own directory. What the
// repository server brings - a command into a volume repository, a volume on
// another node than the repository's - is not refused: its runs fail, saying
// why, until then.
func (uc *UC) checkDataBackup(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	job *entity.SchedJob,
) error {
	if err := checkDataBackupApp(scope, job); err != nil {
		return hperrors.Wrap(err)
	}
	if job.DataBackup == nil || job.DataBackup.Source != base.SchedJobDataBackupSourceVolume {
		return nil
	}
	err := uc.dataBackupService.CheckAppVolume(ctx, db, scope.App, job.DataBackup.SourceVolume.ID)
	return hperrors.Wrap(err)
}

// checkDataBackupApp refuses a data backup whose app is not the scope's.
func checkDataBackupApp(scope *entity.ObjectScope, job *entity.SchedJob) error {
	if job.JobType != base.SchedJobTypeDataBackup {
		return nil
	}
	if job.App.ID != scope.AppID {
		return hperrors.NewArgumentInvalid("app").WithExtraDetail("a data backup runs in its own app")
	}
	return nil
}
