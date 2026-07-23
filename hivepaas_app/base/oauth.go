package base

type OAuthKind string

const (
	OAuthKindGithub          OAuthKind = "github"
	OAuthKindGithubApp       OAuthKind = "github-app"
	OAuthKindGitlab          OAuthKind = "gitlab"
	OAuthKindGitea           OAuthKind = "gitea"
	OAuthKindGoogle          OAuthKind = "google"
	OAuthKindMicrosoftOnline OAuthKind = "microsoft-online"
	OAuthKindOpenIDConnect   OAuthKind = "openid-connect"
)

var (
	AllOAuthKinds = []OAuthKind{OAuthKindGithub, OAuthKindGithubApp, OAuthKindGitlab, OAuthKindGitea,
		OAuthKindGoogle, OAuthKindMicrosoftOnline, OAuthKindOpenIDConnect}
)

const (
	OAuthScopeDefaultGitea           = "read:user user:email"
	OAuthScopeDefaultGithub          = "read:user user:email"
	OAuthScopeDefaultGitlab          = "read_user"
	OAuthScopeDefaultGoogle          = "openid email profile"
	OAuthScopeDefaultMicrosoftOnline = "openid profile email User.Read"
	OAuthScopeDefaultOpenIDConnect   = "openid profile email"
)
