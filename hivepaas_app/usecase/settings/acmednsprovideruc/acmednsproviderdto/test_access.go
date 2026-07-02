package acmednsproviderdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type TestProviderAccessReq struct {
	*AcmeDnsProviderBaseReq
	TestDomain string `json:"testDomain"`
}

func NewTestProviderAccessReq() *TestProviderAccessReq {
	return &TestProviderAccessReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *TestProviderAccessReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	validators = append(validators, basedto.ValidateStr(&req.TestDomain, true, 1, base.DomainNameMaxLen,
		"testDomain")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type TestProviderAccessResp struct {
	Meta *basedto.Meta `json:"meta"`
}
