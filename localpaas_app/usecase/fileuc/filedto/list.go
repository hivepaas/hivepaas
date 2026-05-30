package filedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

type ListFileReq struct {
	Types        []base.FileType        `json:"-" mapstructure:"type"`
	Statuses     []base.SettingStatus   `json:"-" mapstructure:"status"`
	Keys         []string               `json:"-" mapstructure:"key"`
	StorageTypes []base.FileStorageType `json:"-" mapstructure:"storageType"`
	Search       string                 `json:"-" mapstructure:"search"`

	Paging basedto.Paging `json:"-"`
}

func NewListFileReq() *ListFileReq {
	return &ListFileReq{
		Paging: basedto.Paging{
			// Default paging if unset by client
			Sort: basedto.Orders{{Direction: basedto.DirectionDesc, ColumnName: "name"}},
		},
	}
}

func (req *ListFileReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	// TODO: add validation
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListFileResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*FileResp       `json:"data"`
}

func TransformFiles(files []*entity.File) (resp []*FileResp, err error) {
	resp = make([]*FileResp, 0, len(files))
	for _, file := range files {
		item, err := TransformFile(file)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
