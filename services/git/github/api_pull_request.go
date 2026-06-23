package github

import (
	"context"
	"net/http"

	gogithub "github.com/google/go-github/v85/github"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type ListPullRequestOption func(options *gogithub.PullRequestListOptions)

func (c *Client) ListPullRequest(
	ctx context.Context,
	owner string,
	repo string,
	paging *basedto.Paging,
	options ...ListPullRequestOption,
) ([]*gogithub.PullRequest, *basedto.PagingMeta, error) {
	opts, maxItems := createListOpts(paging)
	if maxItems > 0 && maxItems > MaxListPageSize {
		return c.ListAllPullRequests(ctx, owner, repo, paging, options...)
	}

	listOpts := &gogithub.PullRequestListOptions{
		ListOptions: *opts,
	}
	for _, option := range options {
		option(listOpts)
	}

	output, _, err := c.client.PullRequests.List(ctx, owner, repo, listOpts)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}
	return output, &basedto.PagingMeta{
		Offset: opts.Page * opts.PerPage,
		Limit:  opts.PerPage,
		Total:  -1,
	}, nil
}

func (c *Client) ListAllPullRequests(
	ctx context.Context,
	owner string,
	repo string,
	paging *basedto.Paging,
	options ...ListPullRequestOption,
) ([]*gogithub.PullRequest, *basedto.PagingMeta, error) {
	opts, maxItems := createListOpts(paging)
	listOpts := &gogithub.PullRequestListOptions{
		ListOptions: *opts,
	}
	for _, option := range options {
		option(listOpts)
	}

	var output []*gogithub.PullRequest
	client := c.client
	for {
		result, resp, err := client.PullRequests.List(ctx, owner, repo, listOpts)
		if err != nil {
			return nil, nil, apperrors.Wrap(err)
		}
		output = append(output, result...)
		if resp.NextPage <= 0 || listOpts.Page == resp.NextPage || resp.Rate.Remaining <= 0 {
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
) (*gogithub.PullRequest, error) {
	output, resp, err := c.client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, apperrors.NewNotFound(apperrors.Fmt("PR #%v", number))
	}
	return output, nil
}
