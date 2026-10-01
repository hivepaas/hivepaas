package appdeploymentservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type AppDeploymentReq struct {
	*queue.TaskExecData
}

type AppDeploymentResp struct {
	Deployment *entity.Deployment
}

// DeploymentArgs is what one deployment asked for, as opposed to what the app is
// configured with. It travels in the deploy task's arguments and is never stored
// on the app.
type DeploymentArgs struct {
	NoCache bool
	// ImageTags carry no environment prefix; the build adds it.
	ImageTags []string
}

// CheckBuildSourceReq is what a build from a repository or from a function's
// code is about to be configured with.
type CheckBuildSourceReq struct {
	// RepoSource is the repository the build checks out, nil for none.
	RepoSource     *entity.DeploymentRepoSource
	PushToRegistry entity.ObjectID
	// RefObjects hold the settings the repository source refers to: its
	// credentials.
	RefObjects *entity.RefObjects
}
