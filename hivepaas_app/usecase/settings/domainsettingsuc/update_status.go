package domainsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/domainsettingsuc/domainsettingsdto"
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
