package oauthdto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// baseErrors are the fields the request's own validation refuses.
func baseErrors(req *OAuthBaseReq) []string {
	var fields []string
	for _, err := range vld.Validate(req.validate("")...) {
		fields = append(fields, err.Field().Name)
	}
	return fields
}

func oauthReq(kind base.OAuthKind) *OAuthBaseReq {
	return &OAuthBaseReq{Kind: kind, Name: "Sign-in", ClientID: "client-id", ClientSecret: "client-secret-value"}
}

// Only an OpenID Connect provider's emails can be trusted by its setting; every other kind keeps the flag off.
func TestTrustEmailIsForOpenIDConnectOnly(t *testing.T) {
	oidc := oauthReq(base.OAuthKindOpenIDConnect)
	oidc.TrustEmail = true
	require.NoError(t, oidc.modifyRequest())
	assert.True(t, oidc.ToEntity().TrustEmail)

	for _, kind := range []base.OAuthKind{base.OAuthKindGoogle, base.OAuthKindGithub, base.OAuthKindGitlab,
		base.OAuthKindGitea, base.OAuthKindMicrosoftOnline} {
		req := oauthReq(kind)
		req.TrustEmail = true
		require.NoError(t, req.modifyRequest())
		assert.False(t, req.ToEntity().TrustEmail, kind)
	}
}

// A GitLab or Gitea of one's own takes its three URLs together: one alone would be ignored at sign-in.
func TestCustomURLsComeTogether(t *testing.T) {
	const host = "https://git.acme.example"
	for _, kind := range []base.OAuthKind{base.OAuthKindGitlab, base.OAuthKindGitea} {
		partial := oauthReq(kind)
		partial.AuthURL = host + "/oauth/authorize"
		assert.Len(t, baseErrors(partial), 2, kind)

		full := oauthReq(kind)
		full.AuthURL, full.TokenURL, full.ProfileURL = host+"/oauth/authorize", host+"/oauth/token", host+"/api/v4/user"
		assert.Empty(t, baseErrors(full), kind)

		assert.Empty(t, baseErrors(oauthReq(kind)), "gitlab.com or gitea.com")
	}
}
