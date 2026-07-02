package gitcredentialdto

import (
	gogitea "code.gitea.io/sdk/gitea"
	"github.com/google/go-github/v85/github"
	vld "github.com/tiendc/go-validator"
	gogitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type ListBranchReq struct {
	settings.GetSettingReq
	Owner  string         `json:"-" mapstructure:"owner"`
	Repo   string         `json:"-" mapstructure:"repo"`
	Paging basedto.Paging `json:"-"`
}

func NewListBranchReq() *ListBranchReq {
	return &ListBranchReq{}
}

func (req *ListBranchReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.Owner, false, 1, nameMaxLen, "owner")...)
	validators = append(validators, basedto.ValidateStr(&req.Repo, true, 1, nameMaxLen, "repo")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListBranchResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*BranchResp     `json:"data"`
}

type BranchResp struct {
	Name string `json:"name"`
	SHA  string `json:"sha"`
	Ref  string `json:"ref"`
}

func TransformGithubBranch(br *github.Branch) (resp *BranchResp, err error) {
	resp = &BranchResp{
		Name: br.GetName(),
		Ref:  "refs/heads/" + br.GetName(),
	}
	if br.Commit != nil {
		resp.SHA = br.Commit.GetSHA()
	}
	return resp, nil
}

func TransformGithubBranches(branches []*github.Branch) ([]*BranchResp, error) {
	resp, err := basedto.TransformObjectSlice(branches, TransformGithubBranch)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return resp, nil
}

func TransformGitlabBranch(br *gogitlab.Branch) (resp *BranchResp, err error) {
	resp = &BranchResp{
		Name: br.Name,
		Ref:  "refs/heads/" + br.Name,
	}
	if br.Commit != nil {
		resp.SHA = br.Commit.ID
	}
	return resp, nil
}

func TransformGitlabBranches(branches []*gogitlab.Branch) ([]*BranchResp, error) {
	resp, err := basedto.TransformObjectSlice(branches, TransformGitlabBranch)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return resp, nil
}

func TransformGiteaBranch(br *gogitea.Branch) (resp *BranchResp, err error) {
	resp = &BranchResp{
		Name: br.Name,
		Ref:  "refs/heads/" + br.Name,
	}
	if br.Commit != nil {
		resp.SHA = br.Commit.ID
	}
	return resp, nil
}

func TransformGiteaBranches(branches []*gogitea.Branch) ([]*BranchResp, error) {
	resp, err := basedto.TransformObjectSlice(branches, TransformGiteaBranch)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return resp, nil
}
