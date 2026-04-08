package githubappuc

import (
	"github.com/localpaas/localpaas/localpaas_app/repository/cacherepository"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
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
