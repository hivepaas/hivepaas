package imagebuildservice

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type ImageBuildReq struct {
	*queue.TaskExecData
	App *entity.App

	CommitHash     string
	Dockerfile     entity.DeploymentDockerfile
	ImageName      string
	PushToRegistry entity.ObjectID

	ImageBuildSettings *entity.ImageBuildSettings
	NoCache            bool
	BuildID            string

	CheckoutDir string
	TempDir     string // can be empty
}

type ImageBuildResp struct {
	ImageTags []string
}

type BuildNodeResp struct {
	Node            *swarm.Node
	CurrentNodeID   string
	ReleaseNodeFunc func()
}

func (resp *BuildNodeResp) ReleaseNode() {
	if resp.ReleaseNodeFunc != nil {
		resp.ReleaseNodeFunc()
		resp.ReleaseNodeFunc = nil
	}
}
