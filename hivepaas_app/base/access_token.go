package base

type AccessTokenKind string

const (
	AccessTokenKindGithub    AccessTokenKind = "github"
	AccessTokenKindGitlab    AccessTokenKind = "gitlab"
	AccessTokenKindGitea     AccessTokenKind = "gitea"
	AccessTokenKindBitbucket AccessTokenKind = "bitbucket"
	AccessTokenKindGogs      AccessTokenKind = "gogs"
)

var (
	AllGitAccessTokenKinds = []AccessTokenKind{AccessTokenKindGithub, AccessTokenKindGitlab,
		AccessTokenKindGitea, AccessTokenKindBitbucket, AccessTokenKindGogs}

	AllAccessTokenKinds = append([]AccessTokenKind{}, AllGitAccessTokenKinds...)
)
