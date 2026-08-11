package imagebuildserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func (s *service) imageBuild(
	ctx context.Context,
	db database.IDB,
	data *imageBuildData,
) (err error) {
	if data.CheckoutDir == "" {
		return apperrors.NewMissing("CheckoutDir")
	}

	data.ImageTags, err = s.calcBuildImageTags(data.ImageTags, data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Resp.ImageTags = data.ImageTags

	data.EnvVars, err = s.calcBuildEnvVars(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	data.RegistryAuths, err = s.calcBuildRegistryAuths(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	switch data.BuildTool {
	case base.BuildToolDocker:
		return s.buildImageWithDocker(ctx, db, data)
	case base.BuildToolRailpack:
		return s.buildImageWithRailpack(ctx, db, data)
	}

	return nil
}
