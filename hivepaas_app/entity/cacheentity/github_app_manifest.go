package cacheentity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/git/github"
)

type GithubAppManifest struct {
	Manifest    *github.AppManifest `json:"manifest"`
	State       string              `json:"state"`
	Reprovision bool                `json:"reprovision"`
	GithubApp   *entity.Setting     `json:"githubApp"`
}
