package appcloneserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

// checkPlanLimit refuses a clone that would take the installation past its
// plan, when the request asks for the check. It counts under the plan-limit
// lock, which db - the clone task's transaction - holds until the clone is
// committed.
//
// What the clone counts against is what it will be: a function when it is a
// function's clone that takes its deployment settings, an app otherwise.
func (s *service) checkPlanLimit(ctx context.Context, db database.IDB, data *appCloneData) error {
	if !data.CheckPlanLimit {
		return nil
	}
	// Read here rather than taken from the source app, which a caller may have
	// loaded without its kind.
	kinds, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.scope = ?", base.ObjectScopeApp),
		bunex.SelectWhere("setting.object_id = ?", data.SrcApp.ID),
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppKind),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	srcIsFunction := false
	for _, kind := range kinds {
		srcIsFunction = srcIsFunction || entity.IsFunctionKind(kind)
	}
	limit := licenseservice.CloneLimit(srcIsFunction, data.CloneSettings.CloneDeploymentSettings)
	return hperrors.Wrap(s.licenseService.Check(ctx, db, licenseservice.Adding(limit, 1)))
}
