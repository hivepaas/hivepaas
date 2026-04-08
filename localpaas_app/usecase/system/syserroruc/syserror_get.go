package syserroruc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/syserroruc/syserrordto"
)

func (uc *UC) GetSysError(
	ctx context.Context,
	auth *basedto.Auth,
	req *syserrordto.GetSysErrorReq,
) (*syserrordto.GetSysErrorResp, error) {
	appErr, err := uc.appErrorRepo.GetByID(ctx, uc.db, req.ID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := syserrordto.TransformSysError(appErr)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &syserrordto.GetSysErrorResp{
		Data: resp,
	}, nil
}
