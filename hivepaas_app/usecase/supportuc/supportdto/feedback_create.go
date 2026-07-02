package supportdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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

// nolint
func (req *FeedbackBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	// TODO: add validation
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
