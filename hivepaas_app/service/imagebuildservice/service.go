package imagebuildservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	ImageBuild(ctx context.Context, db database.IDB, req *ImageBuildReq) (*ImageBuildResp, error)
	// ResolveBuildInputs reads what a build takes from settings and opens its
	// secrets. It needs the data encryption key, so it runs in the app.
	ResolveBuildInputs(ctx context.Context, db database.IDB, req *ImageBuildReq) (*BuildInputs, error)

	SelectBuildWorkerNode(ctx context.Context, buildSetting *entity.ImageBuildSettings) (BuildNodeResp, error)
}
