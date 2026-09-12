// Package loggingdto is the logging configuration as the API reads and writes it.
//
// It is not entity.Logging on the wire. The entity's credentials are
// EncryptedField, which would marshal as ciphertext; here they are strings the
// response always masks and the request carries in the clear.
package loggingdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetSettingsReq struct {
}

func NewGetSettingsReq() *GetSettingsReq {
	return &GetSettingsReq{}
}

func (req *GetSettingsReq) Validate() hperrors.ValidationErrors {
	return nil
}

type GetSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *SettingsData `json:"data"`
	// SecretMasked says the credentials came back as the placeholder. Sending
	// the placeholder back in an update keeps what is stored.
	SecretMasked bool            `json:"secretMasked,omitempty"`
	Status       *SettingsStatus `json:"status"`
}
