package supportdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

const (
	typeMinLen = 1
	typeMaxLen = 50

	nameMinLen = 1
	nameMaxLen = 100

	descMinLen = 1
	descMaxLen = 10000
)

type CreateFeedbackReq struct {
	*FeedbackBaseReq
}

type FeedbackBaseReq struct {
	Type        string `json:"type"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	Company     string `json:"company"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
}

func (req *FeedbackBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Type, true, typeMinLen, typeMaxLen, field+"type")...)
	res = append(res, basedto.ValidateStr(&req.Name, true, nameMinLen, nameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStr(&req.Subject, true, nameMinLen, nameMaxLen, field+"subject")...)
	res = append(res, basedto.ValidateStr(&req.Description, true, descMinLen, descMaxLen, field+"description")...)
	return res
}

func NewCreateFeedbackReq() *CreateFeedbackReq {
	return &CreateFeedbackReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateFeedbackReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateFeedbackResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
