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
		ECR: &RegistryAuthECRReq{KeyAuth: basedto.ObjectIDReq{ID: "ka1"}}}
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

func TestAnECRCredentialIsItsAddressAndKeyAuth(t *testing.T) {
	assert.Empty(t, invalidFields(ecrReq()))

	withRole := ecrReq()
	withRole.ECR.RoleARN = "arn:aws:iam::123456789012:role/hivepaas-pull"
	assert.Empty(t, invalidFields(withRole))

	for name, c := range map[string]struct {
		mutate func(*RegistryAuthBaseReq)
		field  string
	}{
		"an address that is no ECR registry's": {func(r *RegistryAuthBaseReq) { r.Address = "ghcr.io" }, "auth.address"},
		"no ECR side":                          {func(r *RegistryAuthBaseReq) { r.ECR = nil }, "auth.ecr"},
		"no key auth": {func(r *RegistryAuthBaseReq) { r.ECR.KeyAuth.ID = "" },
			"auth.ecr.keyAuth"},
		"a role that is none":   {func(r *RegistryAuthBaseReq) { r.ECR.RoleARN = "admin" }, "auth.ecr.roleArn"},
		"a password of its own": {func(r *RegistryAuthBaseReq) { r.Password = "pw" }, "auth.password"},
		"a kind that is none":   {func(r *RegistryAuthBaseReq) { r.Kind = "gcp" }, "auth.kind"},
	} {
		req := ecrReq()
		c.mutate(req)
		assert.Contains(t, invalidFields(req), c.field, name)
	}

	basic := &RegistryAuthBaseReq{Name: "hub", Address: "docker.io", Username: "bot", Password: "pw",
		ECR: &RegistryAuthECRReq{KeyAuth: basedto.ObjectIDReq{ID: "ka1"}}}
	assert.Contains(t, invalidFields(basic), "auth.ecr", "keys on a username and password credential")
}

func TestAnECRCredentialsRegionIsItsAddresss(t *testing.T) {
	auth := ecrReq().ToEntity()
	assert.Equal(t, "eu-west-1", auth.ECR.Region)
	assert.Equal(t, "ka1", auth.ECR.KeyAuth.ID)
	assert.Equal(t, "AWS", auth.Username)
	assert.True(t, auth.Password.IsEmpty())
	assert.Equal(t, []string{"ka1"}, auth.GetRefObjectIDs().RefSettingIDs)
}

// An edit that leaves the key auth, role and registry as they were keeps the
// token, and the key auth's version it was got with; another key auth starts
// without one.
func TestAnEditKeepsTheTokenWhileItSignsInTheSameWay(t *testing.T) {
	current := ecrReq().ToEntity()
	current.Token = entity.NewEncryptedField("tok")
	current.TokenExpiresAt = time.Now().Add(10 * time.Hour)
	current.TokenKeyVer = 3

	edit := ecrReq()
	edit.Name = "renamed"
	auth := edit.ToEntity()
	KeepToken(auth, current)
	token, _ := auth.Token.GetPlain()
	assert.Equal(t, "tok", token)
	assert.Equal(t, 3, auth.TokenKeyVer)

	relinked := ecrReq()
	relinked.ECR.KeyAuth.ID = "ka2"
	auth = relinked.ToEntity()
	KeepToken(auth, current)
	assert.True(t, auth.Token.IsEmpty(), "another key auth, no token")
}

// The API answers the key auth as a reference and when the token expires:
// never the token, never a key.
func TestTheAnswerNeverHoldsTheToken(t *testing.T) {
	auth := ecrReq().ToEntity()
	auth.Token = entity.NewEncryptedField("tok-secret-value")
	auth.TokenExpiresAt = time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	setting := &entity.Setting{ID: "ra1", Type: base.SettingTypeRegistryAuth, Kind: ecrAddress,
		Status: base.SettingStatusActive}
	setting.MustSetData(auth)
	keyAuth := &entity.Setting{ID: "ka1", Name: "aws-pull", Type: base.SettingTypeKeyAuth,
		Status: base.SettingStatusActive}
	keyAuth.MustSetData(&entity.KeyAuth{KeyID: "AKIAEXAMPLE000000000", SecretKey: entity.NewEncryptedField("s3cret")})
	refs := entity.NewRefObjects()
	refs.RefSettings["ka1"] = keyAuth

	resp, err := TransformRegistryAuth(setting, refs)
	assert.NoError(t, err)
	raw, _ := json.Marshal(resp)
	assert.NotContains(t, string(raw), "tok-secret-value")
	assert.NotContains(t, string(raw), "s3cret")
	assert.NotContains(t, string(raw), "AKIAEXAMPLE")
	assert.Equal(t, "ka1", resp.AWSECR.KeyAuth.ID)
	assert.Equal(t, "aws-pull", resp.AWSECR.KeyAuth.Name)
	assert.Equal(t, "eu-west-1", resp.AWSECR.Region)
	assert.Equal(t, auth.TokenExpiresAt, resp.AWSECR.TokenExpiresAt)
	assert.Equal(t, base.RegistryAuthKindAWSECR, resp.Kind)
}
