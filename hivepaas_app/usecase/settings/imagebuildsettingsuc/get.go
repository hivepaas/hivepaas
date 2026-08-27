package imagebuildsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imagebuildsettingsuc/imagebuildsettingsdto"
)

func (uc *UC) GetImageBuildSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagebuildsettingsdto.GetImageBuildSettingsReq,
) (*imagebuildsettingsdto.GetImageBuildSettingsResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetUniqueSetting(ctx, auth, &req.GetUniqueSettingReq, &settings.GetUniqueSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	setting := resp.Data
	buildSettings := setting.MustAsImageBuildSettings()
	refClusterObjects := entity.NewRefClusterObjects()

	if len(buildSettings.Workers.NodeIDs) > 0 {
		nodeListResp, err := uc.dockerManager.NodeList(ctx)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		refClusterObjects.AddRefNodes(nodeListResp.Items...)
	}

	respData, err := imagebuildsettingsdto.TransformImageBuild(setting, resp.RefObjects, refClusterObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &imagebuildsettingsdto.GetImageBuildSettingsResp{
		Data: respData,
	}, nil
}
