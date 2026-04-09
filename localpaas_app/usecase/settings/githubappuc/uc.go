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
	*settings.BaseUC
	cacheAppManifestRepo cacherepository.GithubAppManifestRepo
}

func New(
	baseUC *settings.BaseUC,
	cacheAppManifestRepo cacherepository.GithubAppManifestRepo,
) *UC {
	return &UC{
		BaseUC:               baseUC,
		cacheAppManifestRepo: cacheAppManifestRepo,
	}
}
