package githubappdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListGithubAppReq struct {
	settings.ListSettingReq
}

func NewListGithubAppReq() *ListGithubAppReq {
	return &ListGithubAppReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListGithubAppReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListGithubAppResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*GithubAppResp  `json:"data"`
}

func TransformGithubApps(
	settings []*entity.Setting,
	input *GithubAppTransformInput,
) (resp []*GithubAppResp, err error) {
	resp = make([]*GithubAppResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformGithubApp(setting, input)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
