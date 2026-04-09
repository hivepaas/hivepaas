package domainsettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/domainsettingsuc/domainsettingsdto"
)

func (uc *UC) DeleteUniqueDomainSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *domainsettingsdto.DeleteUniqueDomainSettingsReq,
) (*domainsettingsdto.DeleteUniqueDomainSettingsResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteUniqueSetting(ctx, &req.DeleteUniqueSettingReq, &settings.DeleteUniqueSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &domainsettingsdto.DeleteUniqueDomainSettingsResp{}, nil
}
