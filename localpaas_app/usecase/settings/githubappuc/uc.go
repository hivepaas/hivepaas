package githubappuc

import (
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	currentSettingType    = base.SettingTypeGithubApp
	currentSettingVersion = entity.CurrentGithubAppVersion
)

type UC struct {
	cacheAppManifestRepo cacherepository.GithubAppManifestRepo

	*settings.BaseUC
}

func New(
	cacheAppManifestRepo cacherepository.GithubAppManifestRepo,

	baseUC *settings.BaseUC,
) *UC {
	return &UC{
		cacheAppManifestRepo: cacheAppManifestRepo,

		BaseUC: baseUC,
	}
}
