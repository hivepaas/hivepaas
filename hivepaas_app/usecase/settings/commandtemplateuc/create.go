package commandtemplateuc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

const (
	scriptLenThreshold = 2048
)

func (uc *UC) CreateCommandTemplate(
	ctx context.Context,
	auth *basedto.Auth,
	req *commandtemplatedto.CreateCommandTemplateReq,
) (*commandtemplatedto.CreateCommandTemplateResp, error) {
	req.Type = currentSettingType
	cmdTemplate := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: cmdTemplate.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			// Calculate upserting script objects if the script is too long
			upsertingScripts := uc.calcUpsertingScriptSettings(pData.Setting, cmdTemplate, nil)
			pData.UpsertingSettings = append(pData.UpsertingSettings, upsertingScripts...)

			pData.Setting.Kind = string(req.Kind)
			err := pData.Setting.SetData(cmdTemplate)
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &commandtemplatedto.CreateCommandTemplateResp{
		Data: resp.Data,
	}, nil
}

func (uc *UC) calcUpsertingScriptSettings(
	cmdSetting *entity.Setting,
	newCmd *entity.CommandTemplate,
	oldCmd *entity.CommandTemplate,
) (scripts []*entity.Setting) {
	timeNow := time.Now()
	var oldScriptID string

	if oldCmd != nil {
		oldScriptID = oldCmd.Script.ID
	}

	if newCmd != nil {
		if len(newCmd.Script.Value) > scriptLenThreshold {
			scriptSetting := newScriptSetting(cmdSetting, timeNow)
			if oldScriptID != "" {
				scriptSetting.ID = oldScriptID
				oldScriptID = ""
			}
			scriptSetting.MustSetData(&entity.Script{Data: newCmd.Script.Value})
			newCmd.Script.ID = scriptSetting.ID
			newCmd.Script.Value = ""
			scripts = append(scripts, scriptSetting)
		}
	}

	if oldScriptID != "" { // the ID is unused, delete the linked script
		scriptSetting := newScriptSetting(cmdSetting, timeNow)
		scriptSetting.ID = oldScriptID
		scriptSetting.DeletedAt = timeNow
		scripts = append(scripts, scriptSetting)
	}

	return scripts
}

func newScriptSetting(jobSetting *entity.Setting, timeNow time.Time) *entity.Setting {
	return &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       jobSetting.Scope,
		ObjectID:    jobSetting.ObjectID,
		Type:        base.SettingTypeScript,
		Status:      base.SettingStatusActive,
		Name:        "script of cmd template: " + jobSetting.Name,
		Inheritable: jobSetting.Inheritable,
		Version:     entity.CurrentScriptVersion,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}
}
