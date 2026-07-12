package acmednsprovideruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/acmednsprovideruc/acmednsproviderdto"
)

func (uc *UC) UpdateAcmeDnsProviderStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *acmednsproviderdto.UpdateAcmeDnsProviderStatusReq,
) (*acmednsproviderdto.UpdateAcmeDnsProviderStatusResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateSettingStatus(ctx, &req.UpdateSettingStatusReq, &settings.UpdateSettingStatusData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &acmednsproviderdto.UpdateAcmeDnsProviderStatusResp{}, nil
}
