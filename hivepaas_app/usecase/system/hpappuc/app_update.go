package hpappuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc/hpappdto"
)

const (
	lockIDSystemVersionUpdate = "lock:sys:version-update"
)

func (uc *UC) UpdateHpApp(
	ctx context.Context,
	_ *basedto.Auth,
	req *hpappdto.UpdateHpAppReq,
) (*hpappdto.UpdateHpAppResp, error) {
	info, err := uc.hpAppService.GetAppReleaseInfo(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}

	var targetVersion *base.ReleaseInfo
	switch {
	case info.Stable != nil && info.Stable.AppVersion == req.TargetVersion:
		targetVersion = &info.Stable.ReleaseInfo
	case info.Beta != nil && info.Beta.AppVersion == req.TargetVersion:
		targetVersion = &info.Beta.ReleaseInfo
	default:
		return nil, apperrors.New(apperrors.ErrUpdateVerMismatched)
	}

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		_, err := uc.lockRepo.GetByID(ctx, db, lockIDSystemVersionUpdate,
			bunex.SelectFor("UPDATE"),
		)
		if err != nil {
			return apperrors.New(err)
		}
		err = uc.hpAppService.UpdateSystemVersion(ctx, db, targetVersion)
		if err != nil {
			return apperrors.New(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &hpappdto.UpdateHpAppResp{}, nil
}
