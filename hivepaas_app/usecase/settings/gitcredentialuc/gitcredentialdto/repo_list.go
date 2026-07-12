package gitcredentialdto

import (
	"strconv"

	gogitea "code.gitea.io/sdk/gitea"
	"github.com/google/go-github/v85/github"
	vld "github.com/tiendc/go-validator"
	gogitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListRepoReq struct {
	settings.GetSettingReq
	Paging basedto.Paging `json:"-"`
}

func NewListRepoReq() *ListRepoReq {
	return &ListRepoReq{}
}

func (req *ListRepoReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListRepoResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*RepoResp       `json:"data"`
}

type RepoResp struct {
	ID            string `json:"id" copy:"-"`
	Name          string `json:"name"`
	FullName      string `json:"fullName"`
	DefaultBranch string `json:"defaultBranch"`
	CloneURL      string `json:"cloneURL"`
	GitURL        string `json:"gitURL"`
}

func TransformGithubRepo(repo *github.Repository) (resp *RepoResp, err error) {
	if err = copier.Copy(&resp, repo); err != nil {
		return nil, apperrors.Wrap(err)
	}
	resp.ID = strconv.FormatInt(*repo.ID, 10)
	return resp, nil
}

func TransformGithubRepos(repos []*github.Repository) ([]*RepoResp, error) {
	resp, err := basedto.TransformObjectSlice(repos, TransformGithubRepo)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}

func TransformGitlabProject(project *gogitlab.Project) (resp *RepoResp, err error) {
	resp = &RepoResp{
		ID:            strconv.FormatInt(project.ID, 10),
		Name:          project.Name,
		FullName:      project.PathWithNamespace,
		CloneURL:      project.HTTPURLToRepo,
		GitURL:        project.SSHURLToRepo,
		DefaultBranch: project.DefaultBranch,
	}
	return resp, nil
}

func TransformGitlabProjects(projects []*gogitlab.Project) ([]*RepoResp, error) {
	resp, err := basedto.TransformObjectSlice(projects, TransformGitlabProject)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}

func TransformGiteaRepo(repo *gogitea.Repository) (resp *RepoResp, err error) {
	resp = &RepoResp{
		ID:            strconv.FormatInt(repo.ID, 10),
		Name:          repo.Name,
		FullName:      repo.FullName,
		CloneURL:      repo.CloneURL,
		GitURL:        repo.SSHURL,
		DefaultBranch: repo.DefaultBranch,
	}
	return resp, nil
}

func TransformGiteaRepos(repos []*gogitea.Repository) ([]*RepoResp, error) {
	resp, err := basedto.TransformObjectSlice(repos, TransformGiteaRepo)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}
