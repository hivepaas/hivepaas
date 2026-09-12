// Package loggingdto is the logging configuration as the API reads and writes it.
//
// It is not entity.Logging on the wire. The entity's credentials are
// EncryptedField, which would marshal as ciphertext; here they are strings the
// response always masks and the request carries in the clear.
package loggingdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type UpdateSettingsReq struct {
	Data *SettingsData `json:"data"`
}

func NewUpdateSettingsReq() *UpdateSettingsReq {
	return &UpdateSettingsReq{}
}

// Validate rejects a request missing its data, and any credential that looks
// like ciphertext already: EncryptedField.Set would store such a value verbatim
// and it could never be decrypted.
func (req *UpdateSettingsReq) Validate() hperrors.ValidationErrors {
	validators := []vld.Validator{
		vld.Must(req.Data != nil).OnError(
			vld.SetField("data", nil),
			vld.SetCustomKey("ERR_VLD_FIELD_REQUIRED"),
		),
	}
	if req.Data == nil {
		return hperrors.NewValidationErrors(vld.Validate(validators...))
	}
	addEndpoint := func(ep *EndpointData, path string) {
		if ep == nil {
			return
		}
		validators = append(validators, basedto.ValidatePlainSecret(&ep.Password, path+".password")...)
		validators = append(validators, basedto.ValidatePlainSecret(&ep.BearerToken, path+".bearerToken")...)
	}
	addEndpoint(req.Data.Backend.Ingest, "data.backend.ingest")
	addEndpoint(req.Data.Backend.Query, "data.backend.query")
	for i := range req.Data.Forwards {
		addEndpoint(&req.Data.Forwards[i].Endpoint, "data.forwards.endpoint")
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSettingsResp struct {
	Meta         *basedto.Meta `json:"meta"`
	Data         *SettingsData `json:"data"`
	SecretMasked bool          `json:"secretMasked,omitempty"`
}
