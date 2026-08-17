package containeragentdto

import "github.com/hivepaas/hivepaas/hivepaas_app/base"

// DownloadFileInput represents input for copying/downloading a file from a container
type DownloadFileInput struct {
	ContainerID       string                     `json:"containerId"`
	Path              string                     `json:"path"`
	IsDir             bool                       `json:"isDir"`
	CompressionFormat base.FileCompressionFormat `json:"compressionFormat"`
}

// DownloadFileOutput represents chunk output of file download
type DownloadFileOutput struct {
	Chunk       []byte `json:"chunk"`
	FileName    string `json:"fileName,omitempty"`
	FileSize    int64  `json:"fileSize,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}

// DownloadFileStream represents the stream interface for sending file chunks
type DownloadFileStream interface {
	Send(*DownloadFileOutput) error
}
