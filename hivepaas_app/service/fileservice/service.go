package fileservice

import (
	"context"
	"io"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
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

	// Open, Create, Remove and Stat reach a file's content wherever it is stored:
	// on the system volume, on a ClusterVolume through the agent of its node, or
	// in the cloud. Nothing else joins a file's path with a directory.
	Open(ctx context.Context, db database.IDB, file *entity.File) (io.ReadCloser, error)
	// Create writes a file on a volume. What is written reaches its place once the
	// writer is closed without error; Abort leaves nothing.
	Create(ctx context.Context, db database.IDB, file *entity.File) (FileWriter, error)
	// Remove removes a file; one already gone is removed.
	Remove(ctx context.Context, db database.IDB, file *entity.File) error
	Stat(ctx context.Context, db database.IDB, file *entity.File) (int64, error)
	// ProjectVolume is the project's default volume, where the files of its apps go.
	ProjectVolume(ctx context.Context, db database.IDB, projectID string) (*entity.Setting, error)
	// CountOnVolume counts the files a volume holds.
	CountOnVolume(ctx context.Context, db database.IDB, volumeID string) (int, error)
}
