package gitea

import (
	"context"

	gogitea "code.gitea.io/sdk/gitea"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ListBranchOption func(options *gogitea.ListRepoBranchesOptions)

func (c *Client) ListBranch(
	ctx context.Context,
	owner string,
	repo string,
	paging *basedto.Paging,
	options ...ListBranchOption,
) ([]*gogitea.Branch, *basedto.PagingMeta, error) {
	opts, maxItems := createListOpts(paging)
	if maxItems > 0 && maxItems > MaxListPageSize {
		return c.ListAllBranches(ctx, owner, repo, paging, options...)
	}

	listOpts := gogitea.ListRepoBranchesOptions{
		ListOptions: *opts,
	}
	for _, option := range options {
		option(&listOpts)
	}

	output, resp, err := c.client.ListRepoBranches(owner, repo, listOpts)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}
	return output, &basedto.PagingMeta{
		Offset: opts.Page * opts.PageSize,
		Limit:  opts.PageSize,
		Total:  resp.LastPage * opts.PageSize,
	}, nil
}

func (c *Client) ListAllBranches(
	ctx context.Context,
	owner string,
	repo string,
	paging *basedto.Paging,
	options ...ListBranchOption,
) ([]*gogitea.Branch, *basedto.PagingMeta, error) {
	opts, maxItems := createListOpts(paging)
	listOpts := gogitea.ListRepoBranchesOptions{
		ListOptions: *opts,
	}
	for _, option := range options {
		option(&listOpts)
	}

	var output []*gogitea.Branch
	client := c.client
	for {
		result, resp, err := client.ListRepoBranches(owner, repo, listOpts)
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
