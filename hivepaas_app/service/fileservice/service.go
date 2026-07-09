package fileservice

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/appentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	GetDownloadURL(ctx context.Context, db database.IDB, auth *basedto.Auth, req *GetDownloadURLReq) (
		*GetDownloadURLResp, error)

	GenerateDownloadToken(userID string, fileID string, requireLogin bool,
		expiration time.Duration) (string, error)
	ParseDownloadToken(token string) (*appentity.FileDownloadTokenClaims, error)

	Upload(ctx context.Context, db database.IDB, req *UploadReq) (*UploadResp, error)

	DeleteFileData(ctx context.Context, req *DeleteDataReq) (*DeleteDataResp, error)
}
