package githubappuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
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
