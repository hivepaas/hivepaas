package gitcredentialdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListGitCredentialReq struct {
	settings.ListSettingReq
}

func NewListGitCredentialReq() *ListGitCredentialReq {
	return &ListGitCredentialReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListGitCredentialReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.ListSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListGitCredentialResp struct {
	Meta *basedto.ListMeta    `json:"meta"`
	Data []*GitCredentialResp `json:"data"`
}

type GitCredentialResp struct {
	*settings.BaseSettingResp
}

func TransformGitCredentials(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*GitCredentialResp, err error) {
	resp = make([]*GitCredentialResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformGitCredential(setting, refObjects)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}

func TransformGitCredential(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *GitCredentialResp, err error) {
	resp = &GitCredentialResp{}
	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}
