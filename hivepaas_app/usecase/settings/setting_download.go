package settings

import "io"

type BaseDownloadDataResp struct {
	ContentType   string            `json:"contentType"`
	ContentLength int64             `json:"contentLength"`
	ExtraHeaders  map[string]string `json:"headers"`
	Content       io.ReadCloser     `json:"content"`
}
