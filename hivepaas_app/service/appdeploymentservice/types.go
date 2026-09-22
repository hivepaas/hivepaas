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
