package getstartedserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func (s *service) Finish(ctx context.Context, db database.IDB) error {
	sysStatus, err := s.systemStatusRepo.Get(ctx, db)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if !shouldFinish(sysStatus.NextStep) {
		config.SetInstallationStep(sysStatus.NextStep)
		return nil
	}
	sysStatus.NextStep = base.InstallationStepNone
	sysStatus.UpdateVer++
	sysStatus.UpdatedAt = timeutil.NowUTC()
	err = s.systemStatusRepo.Upsert(ctx, db, sysStatus,
		entity.SystemStatusUpsertingConflictCols, entity.SystemStatusUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}
	config.SetInstallationStep(base.InstallationStepNone)
	return nil
}

// shouldFinish says whether Finish has a step to clear. Reading the dashboard's
// certificate may finish the step at any time, so it must never clear one that
// is not the Get started card's.
func shouldFinish(step base.InstallationStep) bool {
	return step == base.InstallationStepGetStarted
}
