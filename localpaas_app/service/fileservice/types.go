package fileservice

import (
	"io"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
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
	Scope           *base.ObjectScope
	FileType        base.FileType
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
