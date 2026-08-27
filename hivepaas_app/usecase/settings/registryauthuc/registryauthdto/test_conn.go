package registryauthdto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
		return hperrors.Wrap(err)
	}
	// NOTE: make sure req.Name is not empty to not fail the validation
	req.Name = gofn.Coalesce(req.Name, "x")
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *TestRegistryAuthConnReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type TestRegistryAuthConnResp struct {
	Meta *basedto.Meta `json:"meta"`
}
