package acmednsproviderdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type TestProviderAccessReq struct {
	*AcmeDnsProviderBaseReq
	TestDomain string `json:"testDomain"`
}

func NewTestProviderAccessReq() *TestProviderAccessReq {
	return &TestProviderAccessReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *TestProviderAccessReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.validate("")...)
	validators = append(validators, basedto.ValidateStr(&req.TestDomain, true, 1, base.DomainNameMaxLen,
		"testDomain")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type TestProviderAccessResp struct {
	Meta *basedto.Meta `json:"meta"`
}
