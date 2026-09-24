package hpappuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc/hpappdto"
)

// GetHpAppUpdatePlan says what an update to a published version would do,
// component by component, so it can be read before it is started. Nothing is
// changed.
func (uc *UC) GetHpAppUpdatePlan(
	ctx context.Context,
	_ *basedto.Auth,
	req *hpappdto.GetHpAppUpdatePlanReq,
) (*hpappdto.GetHpAppUpdatePlanResp, error) {
	info, err := uc.hpAppService.GetAppReleaseInfo(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	target, err := findTargetRelease(info, req.TargetVersion)
	if err != nil {
		return nil, err
	}

	plan, err := uc.sysUpdateService.PlanUpdate(ctx, uc.db, &target.ReleaseInfo)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &hpappdto.GetHpAppUpdatePlanResp{
		Data: hpappdto.TransformUpdatePlan(info.Current, &hpappdto.UpdateTargetResp{
			AppVersion:  target.AppVersion,
			Channel:     releaseChannelOf(info, &target.ReleaseInfo),
			ReleaseDate: target.ReleaseDate,
			NotesURL:    target.NotesURL,
		}, plan),
	}, nil
}
