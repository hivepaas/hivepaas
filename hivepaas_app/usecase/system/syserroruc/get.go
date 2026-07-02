package syserroruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/syserroruc/syserrordto"
)

func (uc *UC) GetSysError(
	ctx context.Context,
	auth *basedto.Auth,
	req *syserrordto.GetSysErrorReq,
) (*syserrordto.GetSysErrorResp, error) {
	appErr, err := uc.appErrorRepo.GetByID(ctx, uc.db, req.ID)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp, err := syserrordto.TransformSysError(appErr)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &syserrordto.GetSysErrorResp{
		Data: resp,
	}, nil
}
