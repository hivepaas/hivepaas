package commandpipeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandpipeuc/commandpipedto"
)

func (uc *UC) UpdateCommandPipe(
	ctx context.Context,
	auth *basedto.Auth,
	req *commandpipedto.UpdateCommandPipeReq,
) (*commandpipedto.UpdateCommandPipeResp, error) {
	req.Type = currentSettingType
	commandPipe := req.ToEntity()
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: commandPipe.GetRefObjectIDs(),
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			// Validate that the command templates exist
			_, err := uc.validateCommandTemplates(ctx, db, req.Scope, req.CommandPipeBaseReq)
			if err != nil {
				return hperrors.Wrap(err)
			}
			err = pData.Setting.SetData(commandPipe)
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &commandpipedto.UpdateCommandPipeResp{}, nil
}
