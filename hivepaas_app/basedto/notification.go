package basedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type BaseEventNotificationReq struct {
	Success           ObjectIDReq `json:"success"`
	SuccessUseDefault bool        `json:"successUseDefault"`
	Failure           ObjectIDReq `json:"failure"`
	FailureUseDefault bool        `json:"failureUseDefault"`
}

func (req *BaseEventNotificationReq) ToEntity() *entity.BaseEventNotification {
	if req == nil {
		return nil
	}
	return &entity.BaseEventNotification{
		Success:           entity.ObjectID{ID: req.Success.ID},
		SuccessUseDefault: req.SuccessUseDefault,
		Failure:           entity.ObjectID{ID: req.Failure.ID},
		FailureUseDefault: req.FailureUseDefault,
	}
}

func (req *BaseEventNotificationReq) Validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, ValidateObjectIDReq(&req.Success, false, field+"success")...)
	res = append(res, ValidateObjectIDReq(&req.Failure, false, field+"failure")...)
	return res
}

type BaseEventNotificationResp struct {
	Success           *NamedObjectResp `json:"success"`
	SuccessUseDefault bool             `json:"successUseDefault"`
	Failure           *NamedObjectResp `json:"failure"`
	FailureUseDefault bool             `json:"failureUseDefault"`
}

func TransformBaseEventNotification(
	notif *entity.BaseEventNotification,
	refObjects *entity.RefObjects,
) *BaseEventNotificationResp {
	if notif == nil {
		return nil
	}
	resp := &BaseEventNotificationResp{
		SuccessUseDefault: notif.SuccessUseDefault,
		FailureUseDefault: notif.FailureUseDefault,
	}
	if notif.Success.ID != "" {
		notifSetting := refObjects.RefSettings[notif.Success.ID]
		if notifSetting != nil {
			resp.Success = &NamedObjectResp{
				ID:   notif.Success.ID,
				Name: notifSetting.Name,
			}
		}
	}
	if notif.Failure.ID != "" {
		notifSetting := refObjects.RefSettings[notif.Failure.ID]
		if notifSetting != nil {
			resp.Failure = &NamedObjectResp{
				ID:   notif.Failure.ID,
				Name: notifSetting.Name,
			}
		}
	}
	return resp
}
