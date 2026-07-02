package repocheckoutservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

type RepoCheckoutReq struct {
	Project     *entity.Project
	App         *entity.App
	RepoSource  *entity.DeploymentRepoSource
	NoCache     bool
	CredSetting *entity.Setting

	RefObjects  *entity.RefObjects
	LogStore    *tasklog.Store
	TempDir     string
	CheckoutDir string
}

type RepoCheckoutResp struct {
	CommitHash    string
	CommitMessage string
	CommitTitle   string
	CommitAuthor  string
}
