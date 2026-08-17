package appcontainerdto

import (
	"io"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

const (
	filePathMaxLen = 512
)

type DownloadFileFromContainerReq struct {
	ProjectID         string                     `json:"-" mapstructure:"-"`
	ProjectEnvID      string                     `json:"-" mapstructure:"-"`
	AppID             string                     `json:"-" mapstructure:"-"`
	NodeID            string                     `json:"nodeId" mapstructure:"nodeId"`
	ContainerID       string                     `json:"containerId" mapstructure:"containerId"`
	Path              string                     `json:"path" mapstructure:"path"`
	IsDir             bool                       `json:"isDir" mapstructure:"isDir"`
	CompressionFormat base.FileCompressionFormat `json:"compressionFormat" mapstructure:"compressionFormat"`
}

func NewDownloadFileFromContainerReq() *DownloadFileFromContainerReq {
	return &DownloadFileFromContainerReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DownloadFileFromContainerReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 7) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateID(&req.ContainerID, true, "containerId")...)
	validators = append(validators, basedto.ValidateStr(&req.Path, true, 1, filePathMaxLen, "path")...)
	validators = append(validators, basedto.ValidateStrIn(&req.CompressionFormat, false,
		base.AllFileCompressionFormats, "compressionFormat")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DownloadFileFromContainerResp struct {
	FileName    string
	FileSize    int64
	ContentType string
	Reader      io.ReadCloser
}
