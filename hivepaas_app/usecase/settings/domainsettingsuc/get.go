package domainsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/domainsettingsuc/domainsettingsdto"
)

func (uc *UC) GetDomainSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *domainsettingsdto.GetDomainSettingsReq,
) (*domainsettingsdto.GetDomainSettingsResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetUniqueSetting(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := domainsettingsdto.TransformDomainSettings(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &domainsettingsdto.GetDomainSettingsResp{
		Data: respData,
	}, nil
}
