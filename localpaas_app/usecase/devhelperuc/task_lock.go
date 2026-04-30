package devhelperuc

import (
	"context"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/devhelperuc/devhelperdto"
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
			return apperrors.Wrap(err)
		}

		time.Sleep(req.Duration.ToDuration())
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &devhelperdto.LockTaskResp{}, nil
}
