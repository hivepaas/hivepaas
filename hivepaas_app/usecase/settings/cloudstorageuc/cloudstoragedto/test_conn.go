package cloudstoragedto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type TestCloudStorageConnReq struct {
	*CloudStorageBaseReq
}

func NewTestCloudStorageConnReq() *TestCloudStorageConnReq {
	return &TestCloudStorageConnReq{}
}

func (req *TestCloudStorageConnReq) ModifyRequest() error {
	// NOTE: make sure req.Name is not empty to not fail the validation
	req.Name = gofn.Coalesce(req.Name, "x")
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *TestCloudStorageConnReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type TestCloudStorageConnResp struct {
	Meta *basedto.Meta `json:"meta"`
}
