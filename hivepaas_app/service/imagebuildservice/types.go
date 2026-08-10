package imagebuildservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

type ImageBuildReq struct {
	App *entity.App

	BuildTool      base.BuildTool
	CommitHash     string
	DockerfilePath string
	ImageName      string
	PushToRegistry entity.ObjectID

	ImageBuildSettings *entity.ImageBuildSettings
	NoCache            bool

	BuildID     string
	RefObjects  *entity.RefObjects
	LogStore    *tasklog.Store
	CheckoutDir string
	TempDir     string // can be empty
}

type ImageBuildResp struct {
	ImageTags []string
}
