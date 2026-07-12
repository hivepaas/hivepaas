package registryauthdto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type TestRegistryAuthConnReq struct {
	*RegistryAuthBaseReq
}

func NewTestRegistryAuthConnReq() *TestRegistryAuthConnReq {
	return &TestRegistryAuthConnReq{}
}

func (req *TestRegistryAuthConnReq) ModifyRequest() error {
	err := req.modifyRequest()
	if err != nil {
		return apperrors.Wrap(err)
	}
	// NOTE: make sure req.Name is not empty to not fail the validation
	req.Name = gofn.Coalesce(req.Name, "x")
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *TestRegistryAuthConnReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type TestRegistryAuthConnResp struct {
	Meta *basedto.Meta `json:"meta"`
}
