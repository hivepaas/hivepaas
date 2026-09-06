package acmednsproviderdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func storedCloudflare() *entity.AcmeDnsProvider {
	return &entity.AcmeDnsProvider{
		Cloudflare: &entity.AcmeDnsProviderCloudflare{
			AuthToken: entity.NewEncryptedField("the-real-token"),
		},
	}
}

func TestKeepMaskedSecrets(t *testing.T) {
	t.Run("the placeholder keeps the stored token", func(t *testing.T) {
		req := &AcmeDnsProviderBaseReq{
			Kind:       base.AcmeDnsProviderCloudflare,
			Cloudflare: &AcmeDnsProviderCloudflareReq{AuthToken: basedto.MaskedSecret},
		}
		provider := req.ToEntity()
		req.KeepMaskedSecrets(provider, storedCloudflare())
		assert.Equal(t, "the-real-token", provider.Cloudflare.AuthToken.String())
	})

	t.Run("a real token replaces the stored one", func(t *testing.T) {
		req := &AcmeDnsProviderBaseReq{
			Kind:       base.AcmeDnsProviderCloudflare,
			Cloudflare: &AcmeDnsProviderCloudflareReq{AuthToken: "a-new-token"},
		}
		provider := req.ToEntity()
		req.KeepMaskedSecrets(provider, storedCloudflare())
		assert.Equal(t, "a-new-token", provider.Cloudflare.AuthToken.String())
	})

	// Switching provider leaves nothing to resolve the placeholder against. It must
	// not be resolved against the previous provider's secret, and it must not be
	// stored either - EncryptedField refuses to encrypt it further down.
	t.Run("switching provider does not borrow the previous secret", func(t *testing.T) {
		req := &AcmeDnsProviderBaseReq{
			Kind:    base.AcmeDnsProviderHetzner,
			Hetzner: &AcmeDnsProviderHetznerReq{APIToken: basedto.MaskedSecret},
		}
		provider := req.ToEntity()
		req.KeepMaskedSecrets(provider, storedCloudflare())
		assert.Equal(t, basedto.MaskedSecret, provider.Hetzner.APIToken.String())
	})

	t.Run("a kind with no secret is a no-op", func(t *testing.T) {
		req := &AcmeDnsProviderBaseReq{
			Kind:    base.AcmeDnsProviderAcmeDNS,
			AcmeDNS: &AcmeDnsProviderAcmeDNSReq{APIBase: "https://acme-dns.example"},
		}
		provider := req.ToEntity()
		req.KeepMaskedSecrets(provider, storedCloudflare())
		assert.Equal(t, "https://acme-dns.example", provider.AcmeDNS.APIBase)
	})
}

// Every provider that stores a secret must be reachable, or its placeholder is
// never resolved and updating it destroys the secret.
func TestSecretFieldsCoversEveryProviderWithASecret(t *testing.T) {
	withoutSecret := map[base.AcmeDnsProvider]bool{base.AcmeDnsProviderAcmeDNS: true}

	for _, kind := range base.AllAcmeDnsProviders {
		req := &AcmeDnsProviderBaseReq{
			Kind:         kind,
			Azure:        &AcmeDnsProviderAzureReq{},
			BaiduCloud:   &AcmeDnsProviderBaiduCloudReq{},
			Cloudflare:   &AcmeDnsProviderCloudflareReq{},
			DigitalOcean: &AcmeDnsProviderDigitalOceanReq{},
			GCloud:       &AcmeDnsProviderGCloudReq{},
			GoDaddy:      &AcmeDnsProviderGoDaddyReq{},
			Hetzner:      &AcmeDnsProviderHetznerReq{},
			HuaweiCloud:  &AcmeDnsProviderHuaweiCloudReq{},
			Namecheap:    &AcmeDnsProviderNamecheapReq{},
			RFC2136:      &AcmeDnsProviderRFC2136Req{},
			Route53:      &AcmeDnsProviderRoute53Req{},
			TencentCloud: &AcmeDnsProviderTencentCloudReq{},
			AcmeDNS:      &AcmeDnsProviderAcmeDNSReq{},
		}
		fields := req.SecretFields()
		if withoutSecret[kind] {
			assert.Empty(t, fields, "kind %q stores no secret", kind)
			continue
		}
		assert.Len(t, fields, 1, "kind %q must expose its secret", kind)
		assert.NotNil(t, req.ToEntity().SecretFieldFor(kind),
			"kind %q must expose its secret on the entity too", kind)
	}
}
