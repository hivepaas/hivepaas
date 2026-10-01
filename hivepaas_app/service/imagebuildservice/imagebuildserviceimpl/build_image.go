package imagebuildserviceimpl

import (
	"context"

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

	if data.PushToRegistry.ID != "" && data.Inputs.PushRegistry == nil {
		return hperrors.NewMissing("Registry auth to push image")
	}

	data.ImageTags, err = buildImageReferences(data.App, data.CommitHash, data.ImageTags, pushRegistry(data.Inputs))
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.Resp.ImageTags = data.ImageTags

	err = s.prepareDockerfile(ctx, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if err = checkSecretsNotDeclaredAsArg(data); err != nil {
		return hperrors.Wrap(err)
	}

	return s.buildImageWithDocker(ctx, db, data)
}
