package vulnignoreuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/vulnignoreuc/vulnignoredto"
)

func (uc *UC) UpdateVulnIgnore(
	ctx context.Context,
	auth *basedto.Auth,
	req *vulnignoredto.UpdateVulnIgnoreReq,
) (*vulnignoredto.UpdateVulnIgnoreResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	ignore := req.ToEntity(auth)
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   ignore.NameOf(),
		VerifyingRefIDs: ignore.GetRefObjectIDs(),
		PrepareUpdate: func(
			_ context.Context,
			_ database.Tx,
			_ *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			pData.Setting.ExpireAt = req.ExpireAtOrZero()
			return hperrors.Wrap(pData.Setting.SetData(ignore))
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &vulnignoredto.UpdateVulnIgnoreResp{}, nil
}
