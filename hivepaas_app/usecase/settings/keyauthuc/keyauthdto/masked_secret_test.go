package keyauthdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestKeepMaskedSecrets(t *testing.T) {
	stored := &entity.KeyAuth{
		KeyID:     "AKIAEXAMPLE",
		SecretKey: entity.NewEncryptedField("the-real-secret-key"),
	}

	t.Run("the placeholder keeps the stored secret key", func(t *testing.T) {
		req := &KeyAuthBaseReq{KeyID: "AKIAEXAMPLE", SecretKey: basedto.MaskedSecret}
		keyAuth := req.ToEntity()
		req.KeepMaskedSecrets(keyAuth, stored)
		assert.Equal(t, "the-real-secret-key", keyAuth.SecretKey.String())
	})

	t.Run("a real secret key replaces the stored one", func(t *testing.T) {
		req := &KeyAuthBaseReq{KeyID: "AKIAEXAMPLE", SecretKey: "a-new-secret-key"}
		keyAuth := req.ToEntity()
		req.KeepMaskedSecrets(keyAuth, stored)
		assert.Equal(t, "a-new-secret-key", keyAuth.SecretKey.String())
	})

	// A value that only looks like the placeholder is a real secret key. Treating it
	// as "unchanged" would quietly refuse a secret key the user did choose.
	t.Run("a near miss is not the placeholder", func(t *testing.T) {
		req := &KeyAuthBaseReq{KeyID: "AKIAEXAMPLE", SecretKey: "*****************"}
		keyAuth := req.ToEntity()
		req.KeepMaskedSecrets(keyAuth, stored)
		assert.Equal(t, "*****************", keyAuth.SecretKey.String())
	})

	t.Run("no stored setting leaves the request alone", func(t *testing.T) {
		req := &KeyAuthBaseReq{KeyID: "AKIAEXAMPLE", SecretKey: basedto.MaskedSecret}
		keyAuth := req.ToEntity()
		req.KeepMaskedSecrets(keyAuth, nil)
		assert.Equal(t, basedto.MaskedSecret, keyAuth.SecretKey.String())
	})
}

func TestCreateRejectsTheMaskedSecret(t *testing.T) {
	// Creation has no stored value to resolve the placeholder against, so it is
	// not a meaningful input the way it is on update.
	req := NewCreateKeyAuthReq()
	req.KeyAuthBaseReq = &KeyAuthBaseReq{
		Name: "backups", KeyID: "AKIAEXAMPLE", SecretKey: basedto.MaskedSecret,
	}
	assert.NotEmpty(t, req.Validate(), "the placeholder must be rejected on create")

	req.SecretKey = "a-real-secret-key"
	errs := req.Validate()
	for _, err := range errs {
		assert.NotContains(t, err.Error(), "secretKey", "a real secret key must be accepted")
	}
}

func TestSecretFields(t *testing.T) {
	req := &KeyAuthBaseReq{SecretKey: "p"}
	fields := req.SecretFields()
	assert.Len(t, fields, 1)
	assert.Equal(t, "secretKey", fields[0].Path)
	assert.Same(t, &req.SecretKey, fields[0].Value)
}
