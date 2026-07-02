package syserroruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/syserroruc/syserrordto"
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
			return apperrors.New(err)
		}

		persistingData := &persistingSysErrorData{}
		uc.prepareDeletingSysError(errData, persistingData)

		return uc.persistData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.New(err)
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
		return apperrors.New(err)
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
