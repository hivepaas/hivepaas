package taskuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/taskuc/taskdto"
)

func (uc *UC) UpdateTaskStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *taskdto.UpdateTaskStatusReq,
) (*taskdto.UpdateTaskStatusResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		taskData := &updateTaskData{}
		err := uc.loadTaskDataForUpdateStatus(ctx, db, req, taskData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		uc.prepareUpdatingTaskStatus(req, taskData)
		return uc.persistTaskStatus(ctx, db, taskData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &taskdto.UpdateTaskStatusResp{}, nil
}

type updateTaskData struct {
	Task *entity.Task
}

func (uc *UC) loadTaskDataForUpdateStatus(
	ctx context.Context,
	db database.IDB,
	req *taskdto.UpdateTaskStatusReq,
	data *updateTaskData,
) error {
	task, err := uc.taskRepo.GetByID(ctx, db, "", req.ID,
		bunex.SelectFor("UPDATE OF task"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if req.UpdateVer != task.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}
	data.Task = task

	return nil
}

func (uc *UC) prepareUpdatingTaskStatus(
	req *taskdto.UpdateTaskStatusReq,
	data *updateTaskData,
) {
	timeNow := timeutil.NowUTC()
	task := data.Task
	task.UpdateVer++
	task.UpdatedAt = timeNow

	if req.Status != nil {
		task.Status = *req.Status
	}
}

func (uc *UC) persistTaskStatus(
	ctx context.Context,
	db database.IDB,
	data *updateTaskData,
) error {
	err := uc.taskRepo.Update(ctx, db, data.Task,
		bunex.UpdateColumns("update_ver", "updated_at", "status"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
