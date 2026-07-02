package devhelperuc

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/devhelperuc/devhelperdto"
)

func (uc *UC) LockTask(
	ctx context.Context,
	auth *basedto.Auth,
	req *devhelperdto.LockTaskReq,
) (*devhelperdto.LockTaskResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		_, err := uc.taskRepo.GetByID(ctx, db, "", req.TaskID,
			bunex.SelectFor("UPDATE"),
		)
		if err != nil {
			return apperrors.New(err)
		}

		time.Sleep(req.Duration.ToDuration())
		return nil
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &devhelperdto.LockTaskResp{}, nil
}
