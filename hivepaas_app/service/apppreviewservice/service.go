package apppreviewservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type Service interface {
	GetPreview(ctx context.Context, db database.IDB, appID, repoRef string, extraOpts ...bunex.SelectQueryOption) (
		*entity.App, error)
	GetPreviews(ctx context.Context, db database.IDB, appID string, extraOpts ...bunex.SelectQueryOption) (
		[]*entity.App, error)

	CreatePreview(ctx context.Context, db database.Tx, req *CreatePreviewReq) (*CreatePreviewResp, error)
}
