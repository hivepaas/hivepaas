package commandtemplateuc

import (
	"context"
	"errors"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/strutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

func (uc *UC) CreateCommandTemplateFromTemplate(
	ctx context.Context,
	auth *basedto.Auth,
	req *commandtemplatedto.CreateCommandTemplateFromTemplateReq,
) (*commandtemplatedto.CreateCommandTemplateFromTemplateResp, error) {
	req.Type = currentSettingType
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName: req.Name,
		Version:       currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			cmdSetting, cmdTpl, err := uc.createTemplatedCommand(ctx, req)
			if err != nil {
				return apperrors.Wrap(err)
			}

			// Calculate upserting script objects if the script is too long
			upsertingScripts := uc.calcUpsertingScriptSettings(pData.Setting, cmdTpl, nil)
			pData.UpsertingSettings = append(pData.UpsertingSettings, upsertingScripts...)

			cmdName := req.Name
			if cmdName == "" {
				cmdName, err = uc.calcCommandTemplateName(ctx, db, req.Scope, currentSettingType, cmdSetting.Name)
				if err != nil {
					return apperrors.Wrap(err)
				}
			}

			pData.Setting.Kind = cmdSetting.Kind
			pData.Setting.Name = gofn.Coalesce(req.Name, cmdName)
			err = pData.Setting.SetData(cmdTpl)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &commandtemplatedto.CreateCommandTemplateFromTemplateResp{
		Data: resp.Data,
	}, nil
}

func (uc *UC) createTemplatedCommand(
	ctx context.Context,
	req *commandtemplatedto.CreateCommandTemplateFromTemplateReq,
) (cmdSetting *entity.Setting, cmdTpl *entity.CommandTemplate, err error) {
	cmdSetting, err = uc.commandService.GetCommand(ctx, req.CommandType, req.CommandKind)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	cmdTpl, err = cmdSetting.AsCommandTemplate()
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	return cmdSetting, cmdTpl, nil
}

func (uc *UC) calcCommandTemplateName(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	settingType base.SettingType,
	baseName string,
) (suggestedPipeName string, err error) {
	for i := 1; i <= 10; i++ {
		name := baseName
		if i > 1 {
			name = fmt.Sprintf("%s (%d)", baseName, i)
		}
		err = uc.checkNameConflict(ctx, db, scope, settingType, name)
		if err == nil {
			return name, nil
		}
	}
	return "", apperrors.Wrap(apperrors.ErrUnavailable).
		WithParam("Name", "Command template name space")
}

func (uc *UC) checkNameConflict(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	settingType base.SettingType,
	name string,
) (err error) {
	if name == "" {
		return nil
	}
	setting, err := uc.SettingRepo.GetByName(ctx, db, scope, settingType, name, false,
		bunex.SelectColumns("id", "name"),
	)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if setting != nil {
		return apperrors.NewAlreadyExist(strutil.ToPascalCase(string(settingType))).
			WithMsgLog("%s '%s' already exists", settingType, setting.Name)
	}
	return nil
}
