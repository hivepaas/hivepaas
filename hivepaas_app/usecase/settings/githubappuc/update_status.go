package githubappuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/githubappuc/githubappdto"
)

func (uc *UC) UpdateGithubAppStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *githubappdto.UpdateGithubAppStatusReq,
) (*githubappdto.UpdateGithubAppStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &githubappdto.UpdateGithubAppStatusResp{}, nil
}
