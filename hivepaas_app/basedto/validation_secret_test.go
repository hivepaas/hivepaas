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

func TestIsMaskedSecret(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "the placeholder is recognized", value: MaskedSecret, want: true},
		{name: "a real secret is not", value: "s3cr3t"},
		{name: "an empty value is not", value: ""},
		// The placeholder used to exist in a 16-star form too. Nothing may treat a
		// value of a different length as the placeholder, or a secret genuinely made
		// of stars would be silently discarded on update.
		{name: "a longer run of stars is not", value: "****************"},
		{name: "a shorter run of stars is not", value: "*******"},
		{name: "the placeholder with padding is not", value: " " + MaskedSecret},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMaskedSecret(tt.value); got != tt.want {
				t.Errorf("IsMaskedSecret(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestValidateNotMaskedSecret(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "the placeholder is rejected", value: MaskedSecret, wantErr: true},
		{name: "a real secret is accepted", value: "s3cr3t"},
		{name: "an empty value is left to the length check", value: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := vld.Validate(ValidateNotMaskedSecret(&tt.value, "secret")...)
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected the value to be rejected")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("expected the value to be accepted, got %v", errs)
			}
		})
	}
}

func TestValidateNotMaskedSecretNil(t *testing.T) {
	if got := ValidateNotMaskedSecret(nil, "secret"); got != nil {
		t.Error("a nil value must produce no validator")
	}
}

func TestValidateNoMaskedSecrets(t *testing.T) {
	real1, masked, real2 := "s3cr3t", MaskedSecret, "an0ther"

	errs := vld.Validate(ValidateNoMaskedSecrets([]SecretField{
		{Path: "a", Value: &real1},
		{Path: "b", Value: &masked},
		{Path: "c", Value: &real2},
	})...)
	if len(errs) != 1 {
		t.Fatalf("expected exactly the masked field to be reported, got %v", errs)
	}

	if got := ValidateNoMaskedSecrets(nil); got != nil {
		t.Error("no fields must produce no validator")
	}
	if got := vld.Validate(ValidateNoMaskedSecrets([]SecretField{
		{Path: "a", Value: &real1},
	})...); len(got) > 0 {
		t.Errorf("expected the real secret to be accepted, got %v", got)
	}
}

// A stored secret is not decryptable if the placeholder ever reaches the entity,
// so the two secret checks must not overlap or cancel each other out.
func TestPlainAndMaskedChecksAreIndependent(t *testing.T) {
	value := MaskedSecret
	if errs := vld.Validate(ValidatePlainSecret(&value, "secret")...); len(errs) > 0 {
		t.Errorf("the placeholder is not an encrypted value, got %v", errs)
	}

	value = base.EncryptionSaltPrefix + "abc"
	if errs := vld.Validate(ValidateNotMaskedSecret(&value, "secret")...); len(errs) > 0 {
		t.Errorf("an encrypted value is not the placeholder, got %v", errs)
	}
}
