package base

type CloudProviderKind string

const (
	CloudProviderKindAWS        CloudProviderKind = "aws"
	CloudProviderKindGCP        CloudProviderKind = "gcp"
	CloudProviderKindAzure      CloudProviderKind = "azure"
	CloudProviderKindCloudflare CloudProviderKind = "cloudflare"
)

var (
	AllCloudProviderKinds = []CloudProviderKind{CloudProviderKindAWS, CloudProviderKindGCP, CloudProviderKindAzure,
		CloudProviderKindCloudflare}
)
