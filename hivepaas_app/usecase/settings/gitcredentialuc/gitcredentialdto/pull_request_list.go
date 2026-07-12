package gitcredentialdto

import (
	"strconv"
	"time"

	gogitea "code.gitea.io/sdk/gitea"
	"github.com/google/go-github/v85/github"
	vld "github.com/tiendc/go-validator"
	gogitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	nameMaxLen = 100
)

type ListPullRequestReq struct {
	settings.GetSettingReq
	Owner  string         `json:"-" mapstructure:"owner"`
	Repo   string         `json:"-" mapstructure:"repo"`
	Paging basedto.Paging `json:"-"`
}

func NewListPullRequestReq() *ListPullRequestReq {
	return &ListPullRequestReq{}
}

func (req *ListPullRequestReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStr(&req.Owner, false, 1, nameMaxLen, "owner")...)
	validators = append(validators, basedto.ValidateStr(&req.Repo, true, 1, nameMaxLen, "repo")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListPullRequestResp struct {
	Meta *basedto.ListMeta  `json:"meta"`
	Data []*PullRequestResp `json:"data"`
}

type PullRequestResp struct {
	ID        string    `json:"id"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	Branch    string    `json:"branch"`
	SHA       string    `json:"sha"`
	Ref       string    `json:"ref"`
	Author    string    `json:"author"`
	HTMLURL   string    `json:"htmlURL"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func TransformGithubPullRequest(pr *github.PullRequest) (resp *PullRequestResp, err error) {
	resp = &PullRequestResp{
		ID:      strconv.FormatInt(pr.GetID(), 10),
		Number:  pr.GetNumber(),
		Title:   pr.GetTitle(),
		State:   pr.GetState(),
		Ref:     "refs/pull/" + strconv.Itoa(pr.GetNumber()) + "/head",
		HTMLURL: pr.GetHTMLURL(),
	}
	if pr.Head != nil {
		resp.Branch = pr.Head.GetRef()
		resp.SHA = pr.Head.GetSHA()
	}
	if pr.User != nil {
		resp.Author = pr.User.GetLogin()
	}
	if pr.CreatedAt != nil {
		resp.CreatedAt = pr.CreatedAt.Time
	}
	if pr.UpdatedAt != nil {
		resp.UpdatedAt = pr.UpdatedAt.Time
	}
	return resp, nil
}

func TransformGithubPullRequests(prs []*github.PullRequest) ([]*PullRequestResp, error) {
	resp, err := basedto.TransformObjectSlice(prs, TransformGithubPullRequest)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}

func TransformGitlabMergeRequest(mr *gogitlab.BasicMergeRequest) (resp *PullRequestResp, err error) {
	resp = &PullRequestResp{
		ID:      strconv.FormatInt(mr.ID, 10),
		Number:  int(mr.IID),
		Title:   mr.Title,
		State:   mr.State,
		Branch:  mr.SourceBranch,
		SHA:     mr.SHA,
		Ref:     "refs/merge-requests/" + strconv.Itoa(int(mr.IID)) + "/head",
		HTMLURL: mr.WebURL,
	}
	if mr.Author != nil {
		resp.Author = mr.Author.Username
	}
	if mr.CreatedAt != nil {
		resp.CreatedAt = *mr.CreatedAt
	}
	if mr.UpdatedAt != nil {
		resp.UpdatedAt = *mr.UpdatedAt
	}
	return resp, nil
}

func TransformGitlabMergeRequests(mrs []*gogitlab.BasicMergeRequest) ([]*PullRequestResp, error) {
	resp, err := basedto.TransformObjectSlice(mrs, TransformGitlabMergeRequest)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}

func TransformGiteaPullRequest(pr *gogitea.PullRequest) (resp *PullRequestResp, err error) {
	resp = &PullRequestResp{
		ID:      strconv.FormatInt(pr.ID, 10),
		Number:  int(pr.Index),
		Title:   pr.Title,
		State:   string(pr.State),
		Ref:     "refs/pull/" + strconv.Itoa(int(pr.Index)) + "/head",
		HTMLURL: pr.HTMLURL,
	}
	if pr.Head != nil {
		resp.Branch = pr.Head.Ref
		resp.SHA = pr.Head.Sha
	}
	if pr.Poster != nil {
		resp.Author = pr.Poster.UserName
	}
	if pr.Created != nil {
		resp.CreatedAt = *pr.Created
	}
	if pr.Updated != nil {
		resp.UpdatedAt = *pr.Updated
	}
	return resp, nil
}

func TransformGiteaPullRequests(prs []*gogitea.PullRequest) ([]*PullRequestResp, error) {
	resp, err := basedto.TransformObjectSlice(prs, TransformGiteaPullRequest)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}
