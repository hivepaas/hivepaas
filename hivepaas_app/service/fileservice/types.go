package fileservice

import (
	"io"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

/// DOWNLOAD

type GetDownloadURLReq struct {
	File         *entity.File
	RequireLogin bool
	Expiration   time.Duration
	CloudPresign bool
	URLPath      string // if empty, `download` is used
	ViewInline   bool
}

type GetDownloadURLResp struct {
	URL string
}

/// UPLOAD

type UploadReq struct {
	Items           []*UploadItemReq
	Scope           *entity.ObjectScope
	FileType        base.FileType
	FileKind        base.FileKind
	StorageType     base.FileStorageType
	StorageID       string
	ParallelUploads uint
	SaveToDB        bool
}

type UploadItemReq struct {
	FilePath string
	FileSize int64
	FileData io.ReadCloser
}

type UploadResp struct {
	Files []*entity.File
}

/// DELETE

type DeleteDataReq struct {
	// DB reads what deleting a cloud file needs, its storage's key auth.
	DB         database.IDB
	File       *entity.File
	RetryMax   int
	RetryDelay time.Duration
}

type DeleteDataResp struct {
}

/// ACCESS

// FileWriter is a file being written. Close keeps what was written; Abort
// discards it, and the file is left as it was.
type FileWriter interface {
	io.WriteCloser
	Abort(err error)
}
