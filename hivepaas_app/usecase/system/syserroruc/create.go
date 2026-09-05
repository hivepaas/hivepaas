package syserroruc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/syserroruc/syserrordto"
)

// CreateSysError records an error.
//
// Repeats of an error already recorded in the last few minutes are dropped rather
// than stored again: one dependency going down otherwise writes thousands of
// identical rows, at the moment the database can least afford them. Nothing is lost
// by it - every error is logged where it was rendered, so the store keeps one row per
// distinct failure while the log keeps the volume.
func (uc *UC) CreateSysError(
	ctx context.Context,
	req *syserrordto.CreateSysErrorReq,
) (*syserrordto.CreateSysErrorResp, error) {
	if !uc.floodGuard.allow(errorFingerprint(req.ErrorInfo)) {
		return &syserrordto.CreateSysErrorResp{}, nil
	}

	persistingData := &persistingSysErrorData{}
	uc.preparePersistingSysError(req, persistingData)

	err := uc.persistData(ctx, uc.db, persistingData)
	if err != nil {
		return nil, hperrors.Wrap(err)
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
		return hperrors.Wrap(err)
	}
	err = uc.appErrorRepo.InsertMulti(ctx, db, persistingData.InsertingSysErrors)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
