package appcontainerdto

import (
	"io"
	"mime/multipart"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type UploadFileToContainerReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	File              *multipart.FileHeader      `json:"-" mapstructure:"file"`
	NodeID            string                     `json:"-" mapstructure:"nodeId"`
	ContainerID       string                     `json:"-" mapstructure:"containerId"`
	Path              string                     `json:"-" mapstructure:"path"`
	Extract           bool                       `json:"-" mapstructure:"extract"`
	CompressionFormat base.FileCompressionFormat `json:"-" mapstructure:"compressionFormat"`
	Overwrite         *bool                      `json:"-" mapstructure:"overwrite"`

	FileName    string        `json:"-" mapstructure:"-"`
	FileSize    int64         `json:"-" mapstructure:"-"`
	FileContent io.ReadCloser `json:"-" mapstructure:"-"`
}

func NewUploadFileToContainerReq() *UploadFileToContainerReq {
	return &UploadFileToContainerReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UploadFileToContainerReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 7) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateID(&req.ContainerID, false, "containerId")...)
	validators = append(validators, basedto.ValidateStr(&req.Path, true, 1, filePathMaxLen, "path")...)
	validators = append(validators, basedto.ValidateStrIn(&req.CompressionFormat, false,
		base.AllFileCompressionFormats, "compressionFormat")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UploadFileToContainerResp struct {
	Data *UploadFileToContainerDataResp `json:"data"`
}

type UploadFileToContainerDataResp struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}
