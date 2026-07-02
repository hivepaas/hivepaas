package notificationdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListNotificationReq struct {
	settings.ListSettingReq
}

func NewListNotificationReq() *ListNotificationReq {
	return &ListNotificationReq{
		ListSettingReq: settings.ListSettingReq{
			Paging: basedto.Paging{
				// Default paging if unset by client
				Sort: basedto.Orders{{Direction: basedto.DirectionAsc, ColumnName: "name"}},
			},
		},
	}
}

func (req *ListNotificationReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.ListSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListNotificationResp struct {
	Meta *basedto.ListMeta   `json:"meta"`
	Data []*NotificationResp `json:"data"`
}

func TransformNotifications(
	settings []*entity.Setting,
	refObjects *entity.RefObjects,
) (resp []*NotificationResp, err error) {
	resp = make([]*NotificationResp, 0, len(settings))
	for _, setting := range settings {
		item, err := TransformNotification(setting, refObjects)
		if err != nil {
			return nil, apperrors.New(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
