package imagebuildservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	ImageBuild(ctx context.Context, db database.IDB, req *ImageBuildReq) (*ImageBuildResp, error)
}
