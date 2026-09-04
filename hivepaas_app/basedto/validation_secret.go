package basedto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// ValidatePlainSecret rejects a user-supplied secret that looks like an already
// encrypted value.
//
// entity.EncryptedField.Set stores anything carrying the encryption prefix as
// ciphertext without encrypting it, so such an input would be persisted verbatim
// and could never be decrypted again - every later read of it fails.
func ValidatePlainSecret(value *string, field string) []vld.Validator {
	if value == nil {
		return nil
	}
	return []vld.Validator{
		vld.Must(!strings.HasPrefix(*value, base.EncryptionSaltPrefix)).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_SECRET_INVALID"),
		),
	}
}
