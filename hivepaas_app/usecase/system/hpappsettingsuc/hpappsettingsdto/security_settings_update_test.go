package hpappsettingsdto

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func TestUpdateSecuritySettingsReqValidate(t *testing.T) {
	t.Run("requires the app secret", func(t *testing.T) {
		req := &UpdateSecuritySettingsReq{ReturnSecretsViaAPI: true}
		errs := req.Validate()
		assert.NotEmpty(t, errs, "turning secret retrieval on must not be possible without the app secret")
	})

	t.Run("accepts a generated app secret", func(t *testing.T) {
		// What a development install writes for itself: 32 random bytes as hex.
		// The strength bound used for a *new* secret is 50 characters, and applying
		// it here would lock such an install out of its own settings.
		req := &UpdateSecuritySettingsReq{AppSecret: gofn.RandTokenAsHex(32)}
		assert.Len(t, req.AppSecret, 64)
		assert.Empty(t, req.Validate())
	})

	t.Run("rejects an absurdly long value", func(t *testing.T) {
		req := &UpdateSecuritySettingsReq{AppSecret: strings.Repeat("a", existingAppSecretMaxLen+1)}
		assert.NotEmpty(t, req.Validate())
	})
}

// The request arrives as JSON. A struct tagged for TOML decodes only by Go's
// case-insensitive fallback, which stops working the moment a field is renamed.
func TestSecuritySettingsJSONNames(t *testing.T) {
	req := &UpdateSecuritySettingsReq{}
	assert.NoError(t, json.Unmarshal(
		[]byte(`{"appSecret":"s","returnSecretsViaApi":true,`+
			`"alwaysReturnSecretTypes":["swarm-join-token"]}`), req))
	assert.Equal(t, "s", req.AppSecret)
	assert.True(t, req.ReturnSecretsViaAPI)
	assert.Equal(t, []string{"swarm-join-token"}, req.AlwaysReturnSecretTypes)

	body, err := json.Marshal(&SecuritySettingsResp{
		ReturnSecretsViaAPI:     true,
		AlwaysReturnSecretTypes: []string{"swarm-join-token"},
	})
	assert.NoError(t, err)
	assert.JSONEq(t,
		`{"returnSecretsViaApi":true,"alwaysReturnSecretTypes":["swarm-join-token"]}`,
		string(body))
}

func TestUpdateSecuritySettingsReqRejectsUnknownSecretType(t *testing.T) {
	valid := &UpdateSecuritySettingsReq{
		AppSecret:               gofn.RandTokenAsHex(32),
		AlwaysReturnSecretTypes: []string{string(base.SecretTypeSwarmJoinToken)},
	}
	assert.Empty(t, valid.Validate())

	// A name that is not a secret type must be refused rather than stored: stored,
	// it exempts nothing while the operator believes a door is open.
	invalid := &UpdateSecuritySettingsReq{
		AppSecret:               gofn.RandTokenAsHex(32),
		AlwaysReturnSecretTypes: []string{"swarm-join-tokens"},
	}
	assert.NotEmpty(t, invalid.Validate())
}

// The exemptions are behind the same app secret as the flag they punch a hole in.
func TestUpdateSecuritySettingsReqExemptionsNeedTheAppSecret(t *testing.T) {
	req := &UpdateSecuritySettingsReq{
		AlwaysReturnSecretTypes: []string{string(base.SecretTypeSwarmJoinToken)},
	}
	assert.NotEmpty(t, req.Validate())
}
