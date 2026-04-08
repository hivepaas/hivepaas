package syserroruc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/syserroruc/syserrordto"
)

func (uc *UC) DeleteSysError(
	ctx context.Context,
	auth *basedto.Auth,
	req *syserrordto.DeleteSysErrorReq,
) (*syserrordto.DeleteSysErrorResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		errData := &deleteSysErrorData{}
		err := uc.loadSysErrorDataForDelete(ctx, db, req, errData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingSysErrorData{}
		uc.prepareDeletingSysError(errData, persistingData)

		return uc.persistData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &syserrordto.DeleteSysErrorResp{}, nil
}

type deleteSysErrorData struct {
	SysError *entity.SysError
}

func (uc *UC) loadSysErrorDataForDelete(
	ctx context.Context,
	db database.IDB,
	req *syserrordto.DeleteSysErrorReq,
	data *deleteSysErrorData,
) error {
	appError, err := uc.appErrorRepo.GetByID(ctx, db, req.ID,
		bunex.SelectFor("UPDATE"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.SysError = appError

	return nil
}

func (uc *UC) prepareDeletingSysError(
	data *deleteSysErrorData,
	persistingData *persistingSysErrorData,
) {
	persistingData.DeletingSysErrors = append(persistingData.DeletingSysErrors, data.SysError)
}
