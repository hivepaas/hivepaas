package notificationdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteNotificationReq struct {
	settings.DeleteSettingReq
}

func NewDeleteNotificationReq() *DeleteNotificationReq {
	return &DeleteNotificationReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteNotificationReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteNotificationResp struct {
	Meta *basedto.Meta `json:"meta"`
}
