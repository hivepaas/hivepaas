package domainsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/domainsettingsuc/domainsettingsdto"
)

func (uc *UC) DeleteDomainSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *domainsettingsdto.DeleteDomainSettingsReq,
) (*domainsettingsdto.DeleteDomainSettingsResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteUniqueSetting(ctx, &req.DeleteUniqueSettingReq, &settings.DeleteUniqueSettingData{})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &domainsettingsdto.DeleteDomainSettingsResp{}, nil
}
