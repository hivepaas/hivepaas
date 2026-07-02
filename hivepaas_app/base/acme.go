package base

type AcmeDnsProvider string

const (
	AcmeDnsProviderAcmeDNS      AcmeDnsProvider = "acmedns"
	AcmeDnsProviderAzure        AcmeDnsProvider = "azure"
	AcmeDnsProviderBaiduCloud   AcmeDnsProvider = "baiducloud"
	AcmeDnsProviderCloudflare   AcmeDnsProvider = "cloudflare"
	AcmeDnsProviderDigitalOcean AcmeDnsProvider = "digitalocean"
	AcmeDnsProviderGCloud       AcmeDnsProvider = "gcloud"
	AcmeDnsProviderGoDaddy      AcmeDnsProvider = "godaddy"
	AcmeDnsProviderHetzner      AcmeDnsProvider = "hetzner"
	AcmeDnsProviderHuaweiCloud  AcmeDnsProvider = "huaweicloud"
	AcmeDnsProviderNamecheap    AcmeDnsProvider = "namecheap"
	AcmeDnsProviderRFC2136      AcmeDnsProvider = "rfc2136"
	AcmeDnsProviderRoute53      AcmeDnsProvider = "route53"
	AcmeDnsProviderTencentCloud AcmeDnsProvider = "tencentcloud"
)

var (
	AllAcmeDnsProviders = []AcmeDnsProvider{
		AcmeDnsProviderAcmeDNS,
		AcmeDnsProviderAzure,
		AcmeDnsProviderBaiduCloud,
		AcmeDnsProviderCloudflare,
		AcmeDnsProviderDigitalOcean,
		AcmeDnsProviderGCloud,
		AcmeDnsProviderGoDaddy,
		AcmeDnsProviderHetzner,
		AcmeDnsProviderHuaweiCloud,
		AcmeDnsProviderNamecheap,
		AcmeDnsProviderRFC2136,
		AcmeDnsProviderRoute53,
		AcmeDnsProviderTencentCloud,
	}
)
