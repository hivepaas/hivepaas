package filedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type GetFileReq struct {
	settings.GetSettingReq
}

func NewGetFileReq() *GetFileReq {
	return &GetFileReq{}
}

func (req *GetFileReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetFileResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *FileResp     `json:"data"`
}

type FileResp struct {
	*settings.BaseSettingResp
	StorageType base.FileStorageType      `json:"storageType"`
	Storage     *settings.BaseSettingResp `json:"storage"`
	Bucket      string                    `json:"bucket"`
	Mimetype    string                    `json:"mimetype"`
	Size        int64                     `json:"size"`
	Path        string                    `json:"path"`
}

func TransformFile(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *FileResp, err error) {
	config := setting.MustAsFile()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.Storage, err = settings.TransformSettingBase(refObjects.RefSettings[config.Storage.ID])
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}
