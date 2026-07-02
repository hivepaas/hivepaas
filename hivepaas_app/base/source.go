package base

type RepoType string

const (
	RepoTypeGit RepoType = "git"
)

var (
	AllRepoTypes = []RepoType{RepoTypeGit}
)
