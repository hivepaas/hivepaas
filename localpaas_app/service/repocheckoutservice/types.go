package repocheckoutservice

import (
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

type RepoCheckoutReq struct {
	*queue.TaskExecData
	Project     *entity.Project
	App         *entity.App
	RepoSource  *entity.DeploymentRepoSource
	NoCache     bool
	CredSetting *entity.Setting

	TempDir     string
	CheckoutDir string
}

type RepoCheckoutResp struct {
	CommitHash    string
	CommitMessage string
	CommitTitle   string
	CommitAuthor  string
}
