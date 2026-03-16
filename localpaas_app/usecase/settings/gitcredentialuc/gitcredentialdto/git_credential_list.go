package gitcredentialdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
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

func (req *ListGitCredentialReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
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
			return nil, apperrors.Wrap(err)
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
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}
