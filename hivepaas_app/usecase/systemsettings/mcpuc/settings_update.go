package mcpuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/mcpuc/mcpdto"
)

const (
	currentSettingType = base.SettingTypeMCP
	mcpSettingName     = "MCP server settings"
)

// UpdateMCPSettings stores the switch. The base loads the setting for update,
// creates it on the first save, checks updateVer and records the audit entry.
func (uc *UC) UpdateMCPSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *mcpdto.UpdateMCPSettingsReq,
) (*mcpdto.UpdateMCPSettingsResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	next := req.ToEntity()

	_, err := uc.UpdateUniqueSetting(ctx, &req.UpdateUniqueSettingReq, &settings.UpdateUniqueSettingData{
		Name:    mcpSettingName,
		Version: entity.CurrentMCPSettingsVersion,
		PrepareUpdate: func(
			_ context.Context,
			_ database.Tx,
			_ *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			return hperrors.Wrap(pData.Setting.SetData(next))
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	uc.mu.Lock()
	uc.current, uc.read = *next, timeNow()
	uc.mu.Unlock()
	return &mcpdto.UpdateMCPSettingsResp{}, nil
}
