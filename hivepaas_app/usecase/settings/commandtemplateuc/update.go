package commandtemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

func (uc *UC) UpdateCommandTemplate(
	ctx context.Context,
	auth *basedto.Auth,
	req *commandtemplatedto.UpdateCommandTemplateReq,
) (*commandtemplatedto.UpdateCommandTemplateResp, error) {
	req.Type = currentSettingType
	newCmdTemplate := req.ToEntity()
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: newCmdTemplate.GetRefObjectIDs(),
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			oldCmdTemplate, err := data.Setting.AsCommandTemplate()
			if err != nil {
				return apperrors.Wrap(err)
			}

			// Calculate upserting script objects if the script is too long
			upsertingScripts := uc.calcUpsertingScriptSettings(pData.Setting, newCmdTemplate, oldCmdTemplate)
			pData.UpsertingSettings = append(pData.UpsertingSettings, upsertingScripts...)

			pData.Setting.Kind = string(req.Kind)
			err = pData.Setting.SetData(newCmdTemplate)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &commandtemplatedto.UpdateCommandTemplateResp{}, nil
}
