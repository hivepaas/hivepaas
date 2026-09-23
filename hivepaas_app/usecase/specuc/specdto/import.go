package specdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// ValidateImportReq asks what importing a bundle at a scope would do.
//
// The bundle travels inside the JSON body, base64-encoded, with every call:
// nothing is kept on the server between validate and apply, so a bundle that
// carries secrets is never stored.
type ValidateImportReq struct {
	Scope *entity.ObjectScope `json:"-"`

	// Bundle is the archive export produced, as it was downloaded. JSON carries
	// it as base64.
	Bundle []byte `json:"bundle"`
	// Passphrase opens an encrypted bundle. It is never stored or logged.
	Passphrase string                  `json:"passphrase"`
	Selection  specmodel.Selection     `json:"selection"`
	Options    specmodel.ImportOptions `json:"options"`
}

func NewValidateImportReq() *ValidateImportReq {
	return &ValidateImportReq{}
}

// ModifyRequest defaults the option most people mean: make what exists match.
func (req *ValidateImportReq) ModifyRequest() {
	if req.Options.Existing == "" {
		req.Options.Existing = specmodel.ExistingUpdate
	}
}

func (req *ValidateImportReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateSlice(req.Bundle, false, 1, nil, "bundle")...)
	validators = append(validators, basedto.ValidateStr(&req.Passphrase, false,
		1, passphraseMaxLen, "passphrase")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Options.Existing, true,
		[]specmodel.Existing{specmodel.ExistingUpdate, specmodel.ExistingKeep}, "options.existing")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ValidateImportResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *specmodel.ImportPlan `json:"data"`
}
