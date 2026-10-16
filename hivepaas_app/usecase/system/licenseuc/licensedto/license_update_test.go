package licensedto

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

// modifiedReq is the request as the handler hands it on: decoded, then
// normalized, then validated.
func modifiedReq(t *testing.T, body string) *UpdateLicenseReq {
	t.Helper()
	req := NewUpdateLicenseReq()
	assert.NoError(t, json.Unmarshal([]byte(body), req))
	assert.NoError(t, req.ModifyRequest())
	return req
}

// A PUT with no key is a 400 that never reaches the service, so it can never
// clear or change the license. The dashboard's e2e sends {} to every PUT.
func TestUpdateLicenseReqRequiresAKey(t *testing.T) {
	for name, body := range map[string]string{
		"no body":         `{}`,
		"an empty key":    `{"key":""}`,
		"only whitespace": `{"key":" \n\t\r\n "}`,
	} {
		t.Run(name, func(t *testing.T) {
			errs := modifiedReq(t, body).Validate()

			if assert.NotEmpty(t, errs) {
				info := errs.Build(translation.LangEn)
				assert.Equal(t, http.StatusBadRequest, info.Status)
				assert.Equal(t, "key", info.InnerErrors[0].Path)
			}
		})
	}
}

// The bound is license.MaxKeyTextLen, as the service decodes no longer key, and
// counted after the whitespace is gone.
func TestUpdateLicenseReqBoundsTheKey(t *testing.T) {
	longest := strings.Repeat("A", license.MaxKeyTextLen)

	assert.Empty(t, modifiedReq(t, `{"key":"`+longest+`"}`).Validate())
	assert.Empty(t, modifiedReq(t, `{"key":"`+longest[:100]+`\n  `+longest[100:]+`"}`).Validate(),
		"a key wrapped in a paste is no longer than the key")

	errs := modifiedReq(t, `{"key":"`+longest+`A"}`).Validate()
	if assert.NotEmpty(t, errs) {
		info := errs.Build(translation.LangEn)
		assert.Equal(t, http.StatusBadRequest, info.Status)
		assert.Equal(t, "key", info.InnerErrors[0].Path)
		assert.NotContains(t, info.InnerErrors[0].Cause, longest[:64], "the key is not echoed back")
	}
}

// A key is taken with the whitespace an email client or a file put in it.
func TestUpdateLicenseReqNormalizesTheKey(t *testing.T) {
	req := modifiedReq(t, `{"key":"  eyJh\nbGc\r\n\tiOi  \n"}`)

	assert.Equal(t, "eyJhbGciOi", req.Key)
	assert.Empty(t, req.Validate())
}
