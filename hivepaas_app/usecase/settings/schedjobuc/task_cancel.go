package schedjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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
			return apperrors.Wrap(err)
		}
		canceled, err = uc.taskService.CancelTask(ctx, db, req.TaskID, &req.JobID)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &schedjobdto.CancelSchedJobTaskResp{
		Data: &schedjobdto.CancelSchedJobTaskDataResp{Canceled: canceled},
	}, nil
}
