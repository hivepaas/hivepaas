package github

import (
	"context"

	gogithub "github.com/google/go-github/v85/github"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type ListBranchOption func(options *gogithub.BranchListOptions)

func (c *Client) ListBranch(
	ctx context.Context,
	owner string,
	repo string,
	paging *basedto.Paging,
	options ...ListBranchOption,
) ([]*gogithub.Branch, *basedto.PagingMeta, error) {
	opts, maxItems := createListOpts(paging)
	if maxItems > 0 && maxItems > MaxListPageSize {
		return c.ListAllBranches(ctx, owner, repo, paging, options...)
	}

	listOpts := &gogithub.BranchListOptions{
		ListOptions: *opts,
	}
	for _, option := range options {
		option(listOpts)
	}

	output, _, err := c.client.Repositories.ListBranches(ctx, owner, repo, listOpts)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}
	return output, &basedto.PagingMeta{
		Offset: opts.Page * opts.PerPage,
		Limit:  opts.PerPage,
		Total:  -1,
	}, nil
}

func (c *Client) ListAllBranches(
	ctx context.Context,
	owner string,
	repo string,
	paging *basedto.Paging,
	options ...ListBranchOption,
) ([]*gogithub.Branch, *basedto.PagingMeta, error) {
	opts, maxItems := createListOpts(paging)
	listOpts := &gogithub.BranchListOptions{
		ListOptions: *opts,
	}
	for _, option := range options {
		option(listOpts)
	}

	var output []*gogithub.Branch
	client := c.client
	for {
		result, resp, err := client.Repositories.ListBranches(ctx, owner, repo, listOpts)
		if err != nil {
			return nil, nil, apperrors.New(err)
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
