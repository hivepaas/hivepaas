package domainsettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/domainsettingsuc/domainsettingsdto"
)

func (uc *UC) GetUniqueDomainSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *domainsettingsdto.GetUniqueDomainSettingsReq,
) (*domainsettingsdto.GetUniqueDomainSettingsResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetUniqueSetting(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := domainsettingsdto.TransformDomainSettings(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &domainsettingsdto.GetUniqueDomainSettingsResp{
		Data: respData,
	}, nil
}
