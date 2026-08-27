package sshkeydto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListSSHKeyReq struct {
	settings.ListSettingReq
}

func NewListSSHKeyReq() *ListSSHKeyReq {
	return &ListSSHKeyReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListSSHKeyReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.ListSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListSSHKeyResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*SSHKeyResp     `json:"data"`
}

func TransformSSHKeys(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*SSHKeyResp, err error) {
	resp = make([]*SSHKeyResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformSSHKey(setting, refObjects)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
