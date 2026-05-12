package localpaassettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/localpaassettingsuc/localpaassettingsdto"
)

func (uc *UC) GetLocalPaaSSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *localpaassettingsdto.GetLocalPaaSSettingsReq,
) (*localpaassettingsdto.GetLocalPaaSSettingsResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetUniqueSetting(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := localpaassettingsdto.TransformLocalPaaSSettings(resp.Data, resp.RefObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &localpaassettingsdto.GetLocalPaaSSettingsResp{
		Data: respData,
	}, nil
}
