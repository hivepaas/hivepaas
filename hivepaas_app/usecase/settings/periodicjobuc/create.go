package periodicjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/periodicjobuc/periodicjobdto"
)

func (uc *UC) CreatePeriodicJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *periodicjobdto.CreatePeriodicJobReq,
) (*periodicjobdto.CreatePeriodicJobResp, error) {
	req.Type = currentSettingType
	periodicJob := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: periodicJob.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			pData.Setting.Kind = string(req.Kind)
			if err := pData.Setting.SetData(periodicJob); err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &periodicjobdto.CreatePeriodicJobResp{
		Data: resp.Data,
	}, nil
}
