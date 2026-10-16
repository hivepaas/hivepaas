package licensedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
)

// UpdateLicenseReq is a license key to install, as pasted or as the portal sent
// it.
type UpdateLicenseReq struct {
	Key string `json:"key"`
}

func NewUpdateLicenseReq() *UpdateLicenseReq {
	return &UpdateLicenseReq{}
}

// ModifyRequest implements interface basedto.ReqModifier. A key wrapped by an
// email client or indented in a file is the same key: whitespace in it means
// nothing, so it goes before the key is checked.
func (req *UpdateLicenseReq) ModifyRequest() error {
	req.Key = license.NormalizeKeyText(req.Key)
	return nil
}

// Validate implements interface basedto.ReqValidator.
//
// A key is required, which is also what keeps a PUT with no body, or only
// whitespace, from ever clearing or changing the license. Its length is counted
// in bytes, as license.DecodeKeyText counts it: a key is ASCII, and one too long
// to be a key is refused before anything decodes it.
func (req *UpdateLicenseReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 2) //nolint:mnd
	validators = append(validators, vld.Required(&req.Key).OnError(
		vld.SetField("key", nil),
		vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
	))
	if req.Key != "" {
		validators = append(validators, vld.StrByteLen(&req.Key, 1, license.MaxKeyTextLen).OnError(
			vld.SetField("key", nil),
			vld.SetCustomKey("ERR_VLD_FIELD_LENGTH_INVALID"),
		))
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

// UpdateLicenseResp is the license the key put in force, as GET answers it.
type UpdateLicenseResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *LicenseResp  `json:"data"`
}
