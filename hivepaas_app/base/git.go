package base

type GitSource string

const (
	GitSourceGithub    GitSource = "github"
	GitSourceGitlab    GitSource = "gitlab"
	GitSourceGitea     GitSource = "gitea"
	GitSourceBitbucket GitSource = "bitbucket"
	GitSourceGogs      GitSource = "gogs"
)

var (
	AllGitSources = []GitSource{GitSourceGithub, GitSourceGitlab, GitSourceGitea, GitSourceBitbucket,
		GitSourceGogs}
)
