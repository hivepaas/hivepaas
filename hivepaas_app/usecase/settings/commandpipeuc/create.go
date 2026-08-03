package commandpipeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandpipeuc/commandpipedto"
)

func (uc *UC) CreateCommandPipe(
	ctx context.Context,
	auth *basedto.Auth,
	req *commandpipedto.CreateCommandPipeReq,
) (*commandpipedto.CreateCommandPipeResp, error) {
	req.Type = currentSettingType
	commandPipe := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: commandPipe.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			// Validate that the command templates exist
			_, err := uc.validateCommandTemplates(ctx, db, req.Scope, req.CommandPipeBaseReq)
			if err != nil {
				return apperrors.Wrap(err)
			}
			err = pData.Setting.SetData(commandPipe)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &commandpipedto.CreateCommandPipeResp{
		Data: resp.Data,
	}, nil
}

//nolint:unparam
func (uc *UC) validateCommandTemplates(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	req *commandpipedto.CommandPipeBaseReq,
) ([]*entity.Setting, error) {
	// Validate that the command templates exist
	commandTplIDs := make([]string, 0, 2) //nolint:mnd
	if req.SourceCommand.ID != "" {
		commandTplIDs = append(commandTplIDs, req.SourceCommand.ID)
	}
	if req.TargetCommand.ID != "" {
		commandTplIDs = append(commandTplIDs, req.TargetCommand.ID)
	}

	settings, err := uc.SettingRepo.ListByIDs(ctx, db, scope, commandTplIDs, true)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	for _, commandTplID := range commandTplIDs {
		setting := entityutil.FindByID(settings, commandTplID)
		if setting == nil {
			return nil, apperrors.Wrap(apperrors.ErrSettingNotFound).WithParam("Name", commandTplID)
		}
	}
	return settings, nil
}
