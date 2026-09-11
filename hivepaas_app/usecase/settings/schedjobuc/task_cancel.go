package schedjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
)

func (uc *UC) CancelSchedJobTask(
	ctx context.Context,
	auth *basedto.Auth,
	req *schedjobdto.CancelSchedJobTaskReq,
) (_ *schedjobdto.CancelSchedJobTaskResp, err error) {
	req.Type = currentSettingType
	var canceled bool
	err = transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		_, err = uc.GetSettingByID(ctx, db, &req.BaseSettingReq, req.JobID, false)
		if err != nil {
			return hperrors.Wrap(err)
		}
		// The job's scope, which is also the task's: a job's runs are filed under
		// whatever the job itself is under - see the sched job task builder. The
		// target id below already ties the task to this job; the scope is what
		// stops an id from another project reaching this far to be checked.
		_, canceled, err = uc.taskService.CancelTask(ctx, db, req.Scope, req.TaskID, &req.JobID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &schedjobdto.CancelSchedJobTaskResp{
		Data: &schedjobdto.CancelSchedJobTaskDataResp{Canceled: canceled},
	}, nil
}
