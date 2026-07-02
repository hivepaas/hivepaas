package repowebhookdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListRepoWebhookReq struct {
	settings.ListSettingReq
}

func NewListRepoWebhookReq() *ListRepoWebhookReq {
	return &ListRepoWebhookReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListRepoWebhookReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListRepoWebhookResp struct {
	Meta *basedto.ListMeta  `json:"meta"`
	Data []*RepoWebhookResp `json:"data"`
}

func TransformRepoWebhooks(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*RepoWebhookResp, err error) {
	resp = make([]*RepoWebhookResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformRepoWebhook(setting, refObjects)
		if err != nil {
			return nil, apperrors.New(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
