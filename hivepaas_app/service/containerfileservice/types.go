package containerfileservice

import (
	"io"

	"github.com/moby/moby/api/types/container"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

type StreamFileReq struct {
	Path              string
	IsDir             bool
	CompressionFormat base.FileCompressionFormat

	Content io.ReadCloser
	Stat    *container.PathStat
}

type StreamFileResp struct {
	FileName    string
	FileSize    int64
	ContentType string
	Reader      io.ReadCloser
}
