package filedto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type GetFileReq struct {
	ID    string          `json:"-" mapstructure:"-"`
	Types []base.FileType `json:"-" mapstructure:"type"`
}

func NewGetFileReq() *GetFileReq {
	return &GetFileReq{}
}

func (req *GetFileReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetFileResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *FileResp     `json:"data"`
}

type FileResp struct {
	ID          string                    `json:"id"`
	Type        base.FileType             `json:"type"`
	Status      base.FileStatus           `json:"status"`
	Key         string                    `json:"key"`
	Name        string                    `json:"name"`
	Path        string                    `json:"path"`
	Bucket      string                    `json:"bucket,omitempty"`
	Mimetype    string                    `json:"mimetype"`
	SizeBytes   int64                     `json:"sizeBytes"`
	StorageType base.FileStorageType      `json:"storageType"`
	Storage     *settings.BaseSettingResp `json:"storage,omitempty"`
	UpdateVer   int                       `json:"updateVer"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func TransformFile(file *entity.File) (resp *FileResp, err error) {
	if err = copier.Copy(&resp, file); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.SizeBytes = file.Size

	return resp, nil
}
