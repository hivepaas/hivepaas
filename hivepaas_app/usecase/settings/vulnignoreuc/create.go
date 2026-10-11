package vulnignoreuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/vulnignoreuc/vulnignoredto"
)

func (uc *UC) CreateVulnIgnore(
	ctx context.Context,
	auth *basedto.Auth,
	req *vulnignoredto.CreateVulnIgnoreReq,
) (*vulnignoredto.CreateVulnIgnoreResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	ignore := req.ToEntity(auth)
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		// The name is what it sets aside: one scope sets a vulnerability aside
		// once.
		VerifyingName:   ignore.NameOf(),
		VerifyingRefIDs: ignore.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			_ context.Context,
			_ database.Tx,
			_ *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			pData.Setting.ExpireAt = req.ExpireAtOrZero()
			return hperrors.Wrap(pData.Setting.SetData(ignore))
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &vulnignoredto.CreateVulnIgnoreResp{
		Data: resp.Data,
	}, nil
}
