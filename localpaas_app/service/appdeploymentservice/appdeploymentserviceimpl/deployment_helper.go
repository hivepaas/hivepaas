package appdeploymentserviceimpl

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

func (s *service) loadImageBuildSettings(
	ctx context.Context,
	db database.IDB,
	data *repoDeploymentData,
) error {
	setting, err := s.settingRepo.GetSingle(ctx, db, data.Project.GetObjectScope(),
		base.SettingTypeImageBuildSettings, true)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if setting != nil {
		data.ImageBuildSettings = setting.MustAsImageBuildSettings()
	}
	return nil
}
