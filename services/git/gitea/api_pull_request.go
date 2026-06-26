package gitea

import (
	"context"
	"net/http"

	gogitea "code.gitea.io/sdk/gitea"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type ListPullRequestOption func(options *gogitea.ListPullRequestsOptions)

func (c *Client) ListPullRequest(
	ctx context.Context,
	owner string,
	repo string,
	paging *basedto.Paging,
	options ...ListPullRequestOption,
) ([]*gogitea.PullRequest, *basedto.PagingMeta, error) {
	opts, maxItems := createListOpts(paging)
	if maxItems > 0 && maxItems > MaxListPageSize {
		return c.ListAllPullRequests(ctx, owner, repo, paging, options...)
	}

	listOpts := gogitea.ListPullRequestsOptions{
		ListOptions: *opts,
	}
	for _, option := range options {
		option(&listOpts)
	}

	output, resp, err := c.client.ListRepoPullRequests(owner, repo, listOpts)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}
	return output, &basedto.PagingMeta{
		Offset: opts.Page * opts.PageSize,
		Limit:  opts.PageSize,
		Total:  resp.LastPage * opts.PageSize,
	}, nil
}

func (c *Client) ListAllPullRequests(
	ctx context.Context,
	owner string,
	repo string,
	paging *basedto.Paging,
	options ...ListPullRequestOption,
) ([]*gogitea.PullRequest, *basedto.PagingMeta, error) {
	opts, maxItems := createListOpts(paging)
	listOpts := gogitea.ListPullRequestsOptions{
		ListOptions: *opts,
	}
	for _, option := range options {
		option(&listOpts)
	}

	var output []*gogitea.PullRequest
	client := c.client
	for {
		result, resp, err := client.ListRepoPullRequests(owner, repo, listOpts)
		if err != nil {
			return nil, nil, apperrors.New(err)
		}
		output = append(output, result...)
		if resp.NextPage <= 0 || listOpts.Page == resp.NextPage {
			break
		}
		if maxItems > 0 && len(output) >= maxItems {
			break
		}
		listOpts.Page = resp.NextPage
	}

	pagingMeta := &basedto.PagingMeta{
		Total: len(output),
	}
	if paging != nil {
		pagingMeta.Offset = paging.Offset
		pagingMeta.Limit = paging.Limit
	}
	return output, pagingMeta, nil
}

func (c *Client) GetPullRequestByNumber(
	ctx context.Context,
	owner string,
	repo string,
	number int,
) (*gogitea.PullRequest, error) {
	output, resp, err := c.client.GetPullRequest(owner, repo, int64(number))
	if err != nil {
		return nil, apperrors.New(err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, apperrors.New(apperrors.ErrPullRequestNotFound).WithParam("PullRequest", number)
	}
	return output, nil
}

func (c *Client) CreatePullRequestComment(
	ctx context.Context,
	owner string,
	repo string,
	number int,
	body string,
) (*gogitea.Comment, error) {
	comment, _, err := c.client.CreateIssueComment(owner, repo, int64(number),
		gogitea.CreateIssueCommentOption{Body: body})
	if err != nil {
		return nil, apperrors.New(err)
	}
	return comment, nil
}
