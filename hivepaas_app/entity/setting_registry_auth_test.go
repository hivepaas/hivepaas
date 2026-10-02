package entity

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func TestAnECRAddressGivesItsAccountAndRegion(t *testing.T) {
	for address, want := range map[string][2]string{
		"123456789012.dkr.ecr.eu-west-1.amazonaws.com":          {"123456789012", "eu-west-1"},
		"123456789012.dkr.ecr-fips.us-gov-west-1.amazonaws.com": {"123456789012", "us-gov-west-1"},
		"123456789012.dkr.ecr.cn-north-1.amazonaws.com.cn":      {"123456789012", "cn-north-1"},
		"123456789012.dkr-ecr.us-west-1.on.aws":                 {"123456789012", "us-west-1"},
	} {
		account, region, ok := ParseECRAddress(address)
		assert.True(t, ok, address)
		assert.Equal(t, want, [2]string{account, region}, address)
	}
	for _, address := range []string{"ghcr.io", "1234.dkr.ecr.eu-west-1.amazonaws.com",
		"123456789012.dkr.ecr.eu-west-1.amazonaws.com.evil.test", "public.ecr.aws",
		"123456789012.dkr-ecr.us-west-1.on.aws.evil.test", "evil.test/123456789012.dkr.ecr.eu-west-1.amazonaws.com"} {
		_, _, ok := ParseECRAddress(address)
		assert.False(t, ok, address)
	}
}

func ecrAuth(keyAuthID string) *RegistryAuth {
	return &RegistryAuth{Kind: base.RegistryAuthKindAWSECR, Username: "AWS",
		Address: "123456789012.dkr.ecr.eu-west-1.amazonaws.com",
		ECR:     &RegistryAuthECR{Region: "eu-west-1", KeyAuth: ObjectID{ID: keyAuthID}}}
}

// An ECR credential has no password to hand over: a use that does not go
// through the registry auth service fails rather than send an empty one.
func TestAnECRCredentialHasNoAuthHeaderOfItsOwn(t *testing.T) {
	_, err := ecrAuth("ka1").GenerateAuthHeader()
	assert.Error(t, err)
}

// Its key auth is a reference: linked, it shows in use, and saving checks the
// credential's scope can see it.
func TestAnECRCredentialReferencesItsKeyAuth(t *testing.T) {
	assert.Equal(t, []string{"ka1"}, ecrAuth("ka1").GetRefObjectIDs().RefSettingIDs)
	assert.Empty(t, (&RegistryAuth{Username: "bot"}).GetRefObjectIDs().RefSettingIDs)
}

func TestTheSameECRKeys(t *testing.T) {
	a := ecrAuth("ka1")
	assert.True(t, a.SameECRKeys(ecrAuth("ka1")))
	assert.False(t, a.SameECRKeys(ecrAuth("ka2")), "another key auth")
	moved := ecrAuth("ka1")
	moved.Address = "123456789012.dkr.ecr.us-east-1.amazonaws.com"
	assert.False(t, a.SameECRKeys(moved), "another registry")
	assumed := ecrAuth("ka1")
	assumed.ECR.RoleARN = "arn:aws:iam::123456789012:role/puller"
	assert.False(t, a.SameECRKeys(assumed), "another role")
}

// A spec carries the key auth's id - the key auth is exported on its own - and
// never the token got from it.
func TestASpecNeverCarriesTheECRToken(t *testing.T) {
	useDataKey(t)
	auth := ecrAuth("ka1")
	auth.Token = NewEncryptedField("tok")
	auth.TokenExpiresAt = time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	auth.TokenKeyVer = 3
	SpecPolicyFor(base.SettingTypeRegistryAuth).Strip(auth)

	raw, err := json.Marshal(auth)
	assert.NoError(t, err)
	assert.NotContains(t, string(raw), `"token"`)
	assert.NotContains(t, string(raw), `"tokenExpiresAt"`)
	assert.NotContains(t, string(raw), `"tokenKeyVer"`)
	assert.Contains(t, string(raw), `"keyAuth":{"id":"ka1"`)
}
