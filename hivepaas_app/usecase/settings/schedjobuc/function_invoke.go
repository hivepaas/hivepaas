package schedjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/settinghelper"
)

// checkFunctionInvoke refuses a function's call that is not in its own app, or
// in an app that is not a function.
func (uc *UC) checkFunctionInvoke(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	job *entity.SchedJob,
) error {
	if job.JobType != base.SchedJobTypeFunctionInvoke {
		return nil
	}
	settings, _, err := uc.SettingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppKind),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id = ?", scope.AppID),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	err = checkFunctionInvokeApp(scope, job, settinghelper.FindSettingByType(settings, base.SettingTypeAppKind))
	return hperrors.Wrap(err)
}

// checkFunctionInvokeApp refuses a function's call whose app is not the
// scope's, or whose kind - the scope app's - is not a function.
func checkFunctionInvokeApp(scope *entity.ObjectScope, job *entity.SchedJob, kind *entity.Setting) error {
	if job.JobType != base.SchedJobTypeFunctionInvoke {
		return nil
	}
	if job.App.ID != scope.AppID {
		return hperrors.NewArgumentInvalid("app").WithExtraDetail("a function's call runs in the function")
	}
	if !entity.IsFunctionKind(kind) {
		return hperrors.NewArgumentInvalid("jobType").WithExtraDetail("only a function is called")
	}
	return nil
}
