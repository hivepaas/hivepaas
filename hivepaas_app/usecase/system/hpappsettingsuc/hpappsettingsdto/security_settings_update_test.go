package hpappsettingsdto

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tiendc/gofn"
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
		[]byte(`{"appSecret":"s","returnSecretsViaApi":true}`), req))
	assert.Equal(t, "s", req.AppSecret)
	assert.True(t, req.ReturnSecretsViaAPI)

	body, err := json.Marshal(&SecuritySettingsResp{ReturnSecretsViaAPI: true})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"returnSecretsViaApi":true}`, string(body))
}
