package base

type WebhookKind string

const (
	WebhookKindGithub    = WebhookKind(GitSourceGithub)
	WebhookKindGitlab    = WebhookKind(GitSourceGitlab)
	WebhookKindGitea     = WebhookKind(GitSourceGitea)
	WebhookKindBitbucket = WebhookKind(GitSourceBitbucket)
	WebhookKindGogs      = WebhookKind(GitSourceGogs)
)

var (
	AllWebhookKinds = []WebhookKind{WebhookKindGithub, WebhookKindGitlab, WebhookKindGitea,
		WebhookKindBitbucket, WebhookKindGogs}
)

const (
	DefaultWebhookSecretByteLen = 20 // string length should be double
)
