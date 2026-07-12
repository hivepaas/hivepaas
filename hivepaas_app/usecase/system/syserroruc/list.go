package syserroruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/syserroruc/syserrordto"
)

func (uc *UC) ListSysError(
	ctx context.Context,
	auth *basedto.Auth,
	req *syserrordto.ListSysErrorReq,
) (*syserrordto.ListSysErrorResp, error) {
	listOpts := []bunex.SelectQueryOption{}

	if len(req.Status) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhere("app_error.status IN (?)", bunex.List(req.Status)))
	}
	if len(req.Code) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhere("app_error.code IN (?)", bunex.List(req.Code)))
	}
	if req.Search != "" {
		keyword := bunex.MakeLikeOpStr(req.Search, true)
		listOpts = append(listOpts,
			bunex.SelectWhereGroup(
				bunex.SelectWhere("app_error.code ILIKE ?", keyword),
			),
		)
	}

	settings, paging, err := uc.appErrorRepo.List(ctx, uc.db, &req.Paging, listOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := syserrordto.TransformSysErrors(settings)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &syserrordto.ListSysErrorResp{
		Meta: &basedto.ListMeta{Page: paging},
		Data: resp,
	}, nil
}
