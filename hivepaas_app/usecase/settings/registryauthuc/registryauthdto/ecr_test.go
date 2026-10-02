package registryauthdto

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
)

func TestMain(m *testing.M) {
	key, err := datakey.Generate()
	if err != nil {
		panic(err)
	}
	datakey.SetActive(key)
	os.Exit(m.Run())
}

const ecrAddress = "123456789012.dkr.ecr.eu-west-1.amazonaws.com"

func ecrReq() *RegistryAuthBaseReq {
	return &RegistryAuthBaseReq{Name: "ecr", Kind: base.RegistryAuthKindAWSECR, Address: ecrAddress,
		ECR: &RegistryAuthECRReq{AccessKeyID: "AKIAEXAMPLE000000000", SecretAccessKey: "s3cret"}}
}

func invalidFields(req *RegistryAuthBaseReq) []string {
	var out []string
	for _, err := range vld.Validate(req.validate("auth")...) {
		if f := err.Field(); f != nil {
			out = append(out, f.PathString(true, "."))
		}
	}
	return out
}

func TestAnECRCredentialIsItsAddressAndKeys(t *testing.T) {
	assert.Empty(t, invalidFields(ecrReq()))

	withRole := ecrReq()
	withRole.ECR.RoleARN = "arn:aws:iam::123456789012:role/hivepaas-pull"
	assert.Empty(t, invalidFields(withRole))

	for name, c := range map[string]struct {
		mutate func(*RegistryAuthBaseReq)
		field  string
	}{
		"an address that is no ECR registry's": {func(r *RegistryAuthBaseReq) { r.Address = "ghcr.io" }, "auth.address"},
		"no keys":                              {func(r *RegistryAuthBaseReq) { r.ECR = nil }, "auth.ecr"},
		"a key id that is none": {func(r *RegistryAuthBaseReq) { r.ECR.AccessKeyID = "akia-lower" },
			"auth.ecr.accessKeyId"},
		"no secret": {func(r *RegistryAuthBaseReq) { r.ECR.SecretAccessKey = "" },
			"auth.ecr.secretAccessKey"},
		"a role that is none":   {func(r *RegistryAuthBaseReq) { r.ECR.RoleARN = "admin" }, "auth.ecr.roleArn"},
		"a password of its own": {func(r *RegistryAuthBaseReq) { r.Password = "pw" }, "auth.password"},
		"a kind that is none":   {func(r *RegistryAuthBaseReq) { r.Kind = "gcp" }, "auth.kind"},
	} {
		req := ecrReq()
		c.mutate(req)
		assert.Contains(t, invalidFields(req), c.field, name)
	}

	basic := &RegistryAuthBaseReq{Name: "hub", Address: "docker.io", Username: "bot", Password: "pw",
		ECR: &RegistryAuthECRReq{AccessKeyID: "AKIAEXAMPLE000000000"}}
	assert.Contains(t, invalidFields(basic), "auth.ecr", "keys on a username and password credential")
}

func TestAnECRCredentialsRegionIsItsAddresss(t *testing.T) {
	auth := ecrReq().ToEntity()
	assert.Equal(t, "eu-west-1", auth.ECR.Region)
	assert.Equal(t, "AWS", auth.Username)
	assert.True(t, auth.Password.IsEmpty())
}

// An edit that leaves the keys masked keeps them, and the token with them; new
// keys start without a token.
func TestAnEditKeepsTheKeysAndTokenItDoesNotChange(t *testing.T) {
	current := ecrReq().ToEntity()
	current.Token = entity.NewEncryptedField("tok")
	current.TokenExpiresAt = time.Now().Add(10 * time.Hour)

	edit := ecrReq()
	edit.Name = "renamed"
	edit.ECR.SecretAccessKey = basedto.MaskedSecret
	auth := edit.ToEntity()
	edit.KeepMaskedSecrets(auth, current)
	KeepToken(auth, current)
	secret, _ := auth.ECR.SecretAccessKey.GetPlain()
	assert.Equal(t, "s3cret", secret)
	token, _ := auth.Token.GetPlain()
	assert.Equal(t, "tok", token)

	rotated := ecrReq()
	rotated.ECR.SecretAccessKey = "n3w"
	auth = rotated.ToEntity()
	rotated.KeepMaskedSecrets(auth, current)
	KeepToken(auth, current)
	assert.True(t, auth.Token.IsEmpty(), "new keys, no token")
}

// The API answers the keys - the secret masked - and when the token expires,
// never the token.
func TestTheAnswerNeverHoldsTheToken(t *testing.T) {
	auth := ecrReq().ToEntity()
	auth.Token = entity.NewEncryptedField("tok-secret-value")
	auth.TokenExpiresAt = time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	setting := &entity.Setting{ID: "ra1", Type: base.SettingTypeRegistryAuth, Kind: ecrAddress,
		Status: base.SettingStatusActive}
	setting.MustSetData(auth)
	// Read back as stored: the secrets encrypted, as a GET finds them.
	reread := &entity.Setting{ID: "ra1", Type: base.SettingTypeRegistryAuth, Kind: ecrAddress,
		Status: base.SettingStatusActive, Data: setting.Data}
	resp, err := TransformRegistryAuth(reread, nil)
	assert.NoError(t, err)
	raw, _ := json.Marshal(resp)
	assert.NotContains(t, string(raw), "tok-secret-value")
	assert.NotContains(t, string(raw), "s3cret")
	assert.Equal(t, basedto.MaskedSecret, resp.AWSECR.SecretAccessKey)
	assert.Equal(t, "eu-west-1", resp.AWSECR.Region)
	assert.Equal(t, auth.TokenExpiresAt, resp.AWSECR.TokenExpiresAt)
	assert.Equal(t, base.RegistryAuthKindAWSECR, resp.Kind)
}
