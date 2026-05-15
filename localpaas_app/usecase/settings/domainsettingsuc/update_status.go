package domainsettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/domainsettingsuc/domainsettingsdto"
)

func (uc *UC) UpdateDomainSettingsStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *domainsettingsdto.UpdateDomainSettingsStatusReq,
) (*domainsettingsdto.UpdateDomainSettingsStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateUniqueSettingStatus(ctx, &req.UpdateUniqueSettingStatusReq,
		&settings.UpdateUniqueSettingStatusData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &domainsettingsdto.UpdateDomainSettingsStatusResp{}, nil
}
