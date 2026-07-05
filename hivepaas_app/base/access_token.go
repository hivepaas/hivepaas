package base

type AccessTokenKind string

const (
	// Git
	AccessTokenKindGithub    AccessTokenKind = "github"
	AccessTokenKindGitlab    AccessTokenKind = "gitlab"
	AccessTokenKindGitea     AccessTokenKind = "gitea"
	AccessTokenKindBitbucket AccessTokenKind = "bitbucket"
	AccessTokenKindGogs      AccessTokenKind = "gogs"

	// Cloud providers
	AccessTokenKindCloudflare AccessTokenKind = "cloudflare"
)

var (
	AllGitAccessTokenKinds = []AccessTokenKind{AccessTokenKindGithub, AccessTokenKindGitlab,
		AccessTokenKindGitea, AccessTokenKindBitbucket, AccessTokenKindGogs}

	AllCloudProviderAccessTokenKinds = []AccessTokenKind{AccessTokenKindCloudflare}
)
