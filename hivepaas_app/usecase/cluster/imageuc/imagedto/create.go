package imagedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

const (
	imageIDMaxLen   = 100
	imageNameMaxLen = 200
)

type CreateImageReq struct {
	Name         string              `json:"name"`
	RegistryAuth basedto.ObjectIDReq `json:"registryAuth"`
}

func NewCreateImageReq() *CreateImageReq {
	return &CreateImageReq{}
}

func (req *CreateImageReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.Name, true, 1, imageNameMaxLen, "name")...)
	validators = append(validators, basedto.ValidateObjectIDReq(&req.RegistryAuth, false, "registryAuth")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateImageResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
