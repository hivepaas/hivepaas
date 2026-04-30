package lpappuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/lpappuc/lpappdto"
)

const (
	lockIDSystemVersionUpdate = "lock:sys:version-update"
)

func (uc *UC) UpdateLpApp(
	ctx context.Context,
	_ *basedto.Auth,
	req *lpappdto.UpdateLpAppReq,
) (*lpappdto.UpdateLpAppResp, error) {
	info, err := uc.lpAppService.GetAppReleaseInfo(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	var targetVersion *base.ReleaseInfo
	switch {
	case info.Stable != nil && info.Stable.AppVersion == req.TargetVersion:
		targetVersion = &info.Stable.ReleaseInfo
	case info.Beta != nil && info.Beta.AppVersion == req.TargetVersion:
		targetVersion = &info.Beta.ReleaseInfo
	default:
		return nil, apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		_, err := uc.lockRepo.GetByID(ctx, db, lockIDSystemVersionUpdate,
			bunex.SelectFor("UPDATE"),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		err = uc.lpAppService.UpdateSystemVersion(ctx, db, targetVersion)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &lpappdto.UpdateLpAppResp{}, nil
}
