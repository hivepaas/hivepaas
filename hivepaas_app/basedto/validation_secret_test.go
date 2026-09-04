package basedto

import (
	"testing"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func TestValidatePlainSecret(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "a plain secret is accepted", value: "s3cr3t"},
		{name: "an empty secret is left to the length check", value: ""},
		{
			// EncryptedField.Set would store this verbatim as ciphertext, and every
			// later read of it would fail to decrypt.
			name:    "a value carrying the encryption prefix is rejected",
			value:   base.EncryptionSaltPrefix + "c2FsdA== ZW5jcnlwdGVk",
			wantErr: true,
		},
		{
			name:    "the bare prefix is rejected too",
			value:   base.EncryptionSaltPrefix,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := vld.Validate(ValidatePlainSecret(&tt.value, "secret")...)
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected the value to be rejected")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("expected the value to be accepted, got %v", errs)
			}
		})
	}
}

func TestValidatePlainSecretNil(t *testing.T) {
	if got := ValidatePlainSecret(nil, "secret"); got != nil {
		t.Error("a nil value must produce no validator")
	}
}
