package acmednsprovideruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/acmednsprovideruc/acmednsproviderdto"
)

func (uc *UC) DeleteAcmeDnsProvider(
	ctx context.Context,
	auth *basedto.Auth,
	req *acmednsproviderdto.DeleteAcmeDnsProviderReq,
) (*acmednsproviderdto.DeleteAcmeDnsProviderResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &acmednsproviderdto.DeleteAcmeDnsProviderResp{}, nil
}
