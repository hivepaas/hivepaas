package syserroruc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/syserroruc/syserrordto"
)

func (uc *UC) CreateSysError(
	ctx context.Context,
	req *syserrordto.CreateSysErrorReq,
) (*syserrordto.CreateSysErrorResp, error) {
	persistingData := &persistingSysErrorData{}
	uc.preparePersistingSysError(req, persistingData)

	err := uc.persistData(ctx, uc.db, persistingData)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	createdItem := persistingData.InsertingSysErrors[0]
	return &syserrordto.CreateSysErrorResp{
		Data: &basedto.ObjectIDResp{ID: createdItem.ID},
	}, nil
}

type persistingSysErrorData struct {
	InsertingSysErrors []*entity.SysError
	DeletingSysErrors  []*entity.SysError
}

func (uc *UC) preparePersistingSysError(
	req *syserrordto.CreateSysErrorReq,
	persistingData *persistingSysErrorData,
) {
	timeNow := timeutil.NowUTC()
	appErr := &entity.SysError{
		ID:         gofn.Must(ulid.NewStringULID()),
		Status:     req.ErrorInfo.Status,
		Code:       req.ErrorInfo.Code,
		Detail:     req.ErrorInfo.Detail,
		Cause:      req.ErrorInfo.Cause,
		DebugLog:   req.ErrorInfo.DebugLog,
		StackTrace: req.ErrorInfo.StackTrace,
		CreatedAt:  timeNow,
	}
	persistingData.InsertingSysErrors = append(persistingData.InsertingSysErrors, appErr)
}

func (uc *UC) persistData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingSysErrorData,
) error {
	err := uc.appErrorRepo.DeleteMulti(ctx, db, persistingData.DeletingSysErrors)
	if err != nil {
		return apperrors.Wrap(err)
	}
	err = uc.appErrorRepo.InsertMulti(ctx, db, persistingData.InsertingSysErrors)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
