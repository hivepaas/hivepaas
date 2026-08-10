package imagebuildservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	ImageBuild(ctx context.Context, db database.IDB, req *ImageBuildReq) (*ImageBuildResp, error)

	SelectBuildWorkerNode(ctx context.Context, buildSetting *entity.ImageBuildSettings) (*swarm.Node, error)
}
