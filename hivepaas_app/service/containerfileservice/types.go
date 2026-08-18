package containerfileservice

import (
	"io"

	"github.com/moby/moby/api/types/container"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

type PrepareDownloadStreamReq struct {
	Path              string
	IsDir             bool
	CompressionFormat base.FileCompressionFormat

	Content io.ReadCloser
	Stat    *container.PathStat
}

type PrepareDownloadStreamResp struct {
	FileName    string
	FileSize    int64
	ContentType string
	Reader      io.ReadCloser
}

type PrepareUploadTarStreamReq struct {
	Path              string
	FileName          string
	FileSize          int64
	Extract           bool
	CompressionFormat base.FileCompressionFormat

	Content io.ReadCloser
}

type PrepareUploadTarStreamResp struct {
	DestPath  string
	TarStream io.ReadCloser
}
