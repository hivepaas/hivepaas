package apikeydto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func TestAKeyIsGivenOnlyTheCapabilitiesAKeyMayHave(t *testing.T) {
	withCapabilities := func(capabilities ...base.ResourceCapability) *CreateAPIKeyReq {
		req := newCreateReq(&base.AccessActions{Read: true})
		req.Capabilities = capabilities
		return req
	}
	assert.Empty(t, withCapabilities().Validate(), "none by default")
	assert.Empty(t, withCapabilities(base.ResourceCapSecretReveal).Validate())
	assert.NotEmpty(t, withCapabilities(base.ResourceCapAPIKeyCreate).Validate(), "a key may not mint keys")
	assert.NotEmpty(t, withCapabilities(base.ResourceCapSecretReveal, base.ResourceCapSecretReveal).Validate())
}
