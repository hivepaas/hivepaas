package basicauthdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestKeepMaskedSecrets(t *testing.T) {
	stored := &entity.BasicAuth{
		Username: "admin",
		Password: entity.NewEncryptedField("the-real-password"),
	}

	t.Run("the placeholder keeps the stored password", func(t *testing.T) {
		req := &BasicAuthBaseReq{Username: "admin", Password: basedto.MaskedSecret}
		basicAuth := req.ToEntity()
		req.KeepMaskedSecrets(basicAuth, stored)
		assert.Equal(t, "the-real-password", basicAuth.Password.String())
	})

	t.Run("a real password replaces the stored one", func(t *testing.T) {
		req := &BasicAuthBaseReq{Username: "admin", Password: "a-new-password"}
		basicAuth := req.ToEntity()
		req.KeepMaskedSecrets(basicAuth, stored)
		assert.Equal(t, "a-new-password", basicAuth.Password.String())
	})

	// A value that only looks like the placeholder is a real password. Treating it
	// as "unchanged" would quietly refuse a password the user did choose.
	t.Run("a near miss is not the placeholder", func(t *testing.T) {
		req := &BasicAuthBaseReq{Username: "admin", Password: "****************"}
		basicAuth := req.ToEntity()
		req.KeepMaskedSecrets(basicAuth, stored)
		assert.Equal(t, "****************", basicAuth.Password.String())
	})

	t.Run("no stored setting leaves the request alone", func(t *testing.T) {
		req := &BasicAuthBaseReq{Username: "admin", Password: basedto.MaskedSecret}
		basicAuth := req.ToEntity()
		req.KeepMaskedSecrets(basicAuth, nil)
		assert.Equal(t, basedto.MaskedSecret, basicAuth.Password.String())
	})
}

func TestCreateRejectsTheMaskedSecret(t *testing.T) {
	// Creation has no stored value to resolve the placeholder against, so it is
	// not a meaningful input the way it is on update.
	req := NewCreateBasicAuthReq()
	req.BasicAuthBaseReq = &BasicAuthBaseReq{
		Name: "web", Username: "admin", Password: basedto.MaskedSecret,
	}
	assert.NotEmpty(t, req.Validate(), "the placeholder must be rejected on create")

	req.Password = "a-real-password"
	errs := req.Validate()
	for _, err := range errs {
		assert.NotContains(t, err.Error(), "password", "a real password must be accepted")
	}
}

func TestSecretFields(t *testing.T) {
	req := &BasicAuthBaseReq{Password: "p"}
	fields := req.SecretFields()
	assert.Len(t, fields, 1)
	assert.Equal(t, "password", fields[0].Path)
	assert.Same(t, &req.Password, fields[0].Value)
}
