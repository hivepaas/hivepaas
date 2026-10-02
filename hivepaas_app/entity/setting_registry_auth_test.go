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
	} {
		account, region, ok := ParseECRAddress(address)
		assert.True(t, ok, address)
		assert.Equal(t, want, [2]string{account, region}, address)
	}
	for _, address := range []string{"ghcr.io", "1234.dkr.ecr.eu-west-1.amazonaws.com",
		"123456789012.dkr.ecr.eu-west-1.amazonaws.com.evil.test", "public.ecr.aws"} {
		_, _, ok := ParseECRAddress(address)
		assert.False(t, ok, address)
	}
}

func ecrAuth(secret string) *RegistryAuth {
	return &RegistryAuth{Kind: base.RegistryAuthKindAWSECR, Username: "AWS",
		Address: "123456789012.dkr.ecr.eu-west-1.amazonaws.com",
		ECR: &RegistryAuthECR{Region: "eu-west-1", AccessKeyID: "AKIAEXAMPLE000000000",
			SecretAccessKey: NewEncryptedField(secret)}}
}

// An ECR credential has no password to hand over: a use that does not go
// through the registry auth service fails rather than send an empty one.
func TestAnECRCredentialHasNoAuthHeaderOfItsOwn(t *testing.T) {
	_, err := ecrAuth("s3cret").GenerateAuthHeader()
	assert.Error(t, err)
}

func TestTheSameECRKeys(t *testing.T) {
	useDataKey(t)
	a := ecrAuth("s3cret")
	assert.True(t, a.SameECRKeys(ecrAuth("s3cret")))
	assert.False(t, a.SameECRKeys(ecrAuth("other")), "another secret")
	moved := ecrAuth("s3cret")
	moved.Address = "123456789012.dkr.ecr.us-east-1.amazonaws.com"
	assert.False(t, a.SameECRKeys(moved), "another registry")
	assumed := ecrAuth("s3cret")
	assumed.ECR.RoleARN = "arn:aws:iam::123456789012:role/puller"
	assert.False(t, a.SameECRKeys(assumed), "another role")
}

// A spec carries the keys - as every secret, by the export's mode - and never
// the token got from them.
func TestASpecNeverCarriesTheECRToken(t *testing.T) {
	useDataKey(t)
	auth := ecrAuth("s3cret")
	auth.Token = NewEncryptedField("tok")
	auth.TokenExpiresAt = time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	SpecPolicyFor(base.SettingTypeRegistryAuth).Strip(auth)

	raw, err := json.Marshal(auth)
	assert.NoError(t, err)
	assert.NotContains(t, string(raw), `"token"`)
	assert.NotContains(t, string(raw), `"tokenExpiresAt"`)
	assert.Contains(t, string(raw), `"secretAccessKey"`)
}
