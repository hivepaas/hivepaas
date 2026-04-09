package domainsettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/domainsettingsuc/domainsettingsdto"
)

func (uc *UC) UpdateUniqueDomainSettingsMeta(
	ctx context.Context,
	auth *basedto.Auth,
	req *domainsettingsdto.UpdateUniqueDomainSettingsMetaReq,
) (*domainsettingsdto.UpdateUniqueDomainSettingsMetaResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateUniqueSettingMeta(ctx, &req.UpdateUniqueSettingMetaReq, &settings.UpdateUniqueSettingMetaData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &domainsettingsdto.UpdateUniqueDomainSettingsMetaResp{}, nil
}
