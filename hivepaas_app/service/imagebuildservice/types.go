package imagebuildservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

type ImageBuildReq struct {
	Project            *entity.Project
	App                *entity.App
	RepoSource         *entity.DeploymentRepoSource
	ImageBuildSettings *entity.ImageBuildSettings
	NoCache            bool

	BuildID     string
	RefObjects  *entity.RefObjects
	LogStore    *tasklog.Store
	TempDir     string
	CheckoutDir string
}

type ImageBuildResp struct {
	ImageTags []string
}
