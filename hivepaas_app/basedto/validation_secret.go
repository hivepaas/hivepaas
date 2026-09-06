package basedto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// ValidatePlainSecret rejects a user-supplied secret that looks like an already
// encrypted value.
//
// entity.EncryptedField.Set stores anything carrying one of the encryption
// prefixes as ciphertext without encrypting it, so such an input would be
// persisted verbatim and could never be decrypted again.
func ValidatePlainSecret(value *string, field string) []vld.Validator {
	if value == nil {
		return nil
	}
	for _, prefix := range base.AllEncryptionPrefixes {
		if strings.HasPrefix(*value, prefix) {
			return []vld.Validator{
				vld.Must(false).OnError(
					vld.SetField(field, nil),
					vld.SetCustomKey("ERR_VLD_SECRET_INVALID"),
				),
			}
		}
	}
	return nil
}

// MaskedSecret re-exports base.MaskedSecret so DTOs read the placeholder from
// the layer they already live in. See base for what it is and why it is shared.
const MaskedSecret = base.MaskedSecret

// IsMaskedSecret re-exports base.IsMaskedSecret.
func IsMaskedSecret(value string) bool {
	return base.IsMaskedSecret(value)
}

// SecretField is one secret in a request, paired with the JSON path it arrived
// under so a validation error can name it.
type SecretField struct {
	Path  string
	Value *string
}

// ValidateNotMaskedSecret rejects the placeholder where there is no stored value
// to fall back on, which is creation.
//
// On update the placeholder means "keep what is stored", so it is accepted there
// and resolved by the usecase against the loaded setting; see the KeepMaskedSecrets
// methods on the request types.
func ValidateNotMaskedSecret(value *string, field string) []vld.Validator {
	if value == nil || !IsMaskedSecret(*value) {
		return nil
	}
	return []vld.Validator{
		vld.Must(false).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_SECRET_MASKED"),
		),
	}
}

// ValidateNoMaskedSecrets applies ValidateNotMaskedSecret to every field. Request
// types expose their secrets as a slice so the list lives in one place instead of
// being restated by each caller.
func ValidateNoMaskedSecrets(fields []SecretField) (res []vld.Validator) {
	for _, f := range fields {
		res = append(res, ValidateNotMaskedSecret(f.Value, f.Path)...)
	}
	return res
}
