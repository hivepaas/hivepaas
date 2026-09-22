package imagebuildserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func (s *service) imageBuild(
	ctx context.Context,
	db database.IDB,
	data *imageBuildData,
) (err error) {
	if data.CheckoutDir == "" {
		return hperrors.NewMissing("CheckoutDir")
	}

	var regAuth *entity.RegistryAuth
	if data.PushToRegistry.ID != "" {
		regAuthSetting := data.RefObjects.RefSettings[data.PushToRegistry.ID]
		if regAuthSetting == nil {
			return hperrors.NewMissing("Registry auth to push image")
		}
		regAuth = regAuthSetting.MustAsRegistryAuth()
	}
	data.ImageTags, err = buildImageReferences(data.App, data.CommitHash, data.ImageTags, regAuth)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.Resp.ImageTags = data.ImageTags

	data.EnvVars, err = s.calcBuildEnvVars(ctx, db, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	data.RegistryAuths, err = s.calcBuildRegistryAuths(ctx, db, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = s.prepareDockerfile(ctx, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return s.buildImageWithDocker(ctx, db, data)
}
