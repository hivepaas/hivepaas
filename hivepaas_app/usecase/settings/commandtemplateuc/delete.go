package commandtemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

func (uc *UC) DeleteCommandTemplate(
	ctx context.Context,
	auth *basedto.Auth,
	req *commandtemplatedto.DeleteCommandTemplateReq,
) (*commandtemplatedto.DeleteCommandTemplateResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.DeleteSettingData,
			pData *settings.PersistingSettingDeletionData,
		) error {
			currCmd, err := data.Setting.AsCommandTemplate()
			if err != nil {
				return hperrors.Wrap(err)
			}
			// Calculate deleting script objects if there is any
			upsertingScripts := uc.calcUpsertingScriptSettings(data.Setting, nil, currCmd)
			pData.UpsertingSettings = append(pData.UpsertingSettings, upsertingScripts...)
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &commandtemplatedto.DeleteCommandTemplateResp{}, nil
}
