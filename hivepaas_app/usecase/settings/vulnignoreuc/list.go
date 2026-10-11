package vulnignoreuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/vulnignoreuc/vulnignoredto"
)

func (uc *UC) ListVulnIgnore(
	ctx context.Context,
	auth *basedto.Auth,
	req *vulnignoredto.ListVulnIgnoreReq,
) (*vulnignoredto.ListVulnIgnoreResp, error) {
	req.Type = currentSettingType
	resp, err := uc.ListSetting(ctx, auth, &req.ListSettingReq, &settings.ListSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	respData, err := vulnignoredto.TransformVulnIgnores(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &vulnignoredto.ListVulnIgnoreResp{
		Meta: resp.Meta,
		Data: respData,
	}, nil
}
