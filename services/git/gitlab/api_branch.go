package gitlab

import (
	"context"

	gogitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ListBranchOption func(*gogitlab.ListBranchesOptions)

func (c *Client) ListBranch(
	ctx context.Context,
	pid any,
	paging *basedto.Paging,
	options ...ListBranchOption,
) ([]*gogitlab.Branch, *basedto.PagingMeta, error) {
	opts, maxItems := createListOpts(paging)
	if maxItems > 0 && maxItems > MaxListPageSize {
		return c.ListAllBranches(ctx, pid, paging, options...)
	}

	listOpts := &gogitlab.ListBranchesOptions{
		ListOptions: *opts,
	}
	for _, option := range options {
		option(listOpts)
	}

	output, resp, err := c.client.Branches.ListBranches(pid, listOpts, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, nil, apperrors.New(err)
	}
	return output, &basedto.PagingMeta{
		Offset: int(opts.Page * opts.PerPage),
		Limit:  int(opts.PerPage),
		Total:  int(resp.TotalItems),
	}, nil
}

func (c *Client) ListAllBranches(
	ctx context.Context,
	pid any,
	paging *basedto.Paging,
	options ...ListBranchOption,
) ([]*gogitlab.Branch, *basedto.PagingMeta, error) {
	opts, maxItems := createListOpts(paging)
	listOpts := &gogitlab.ListBranchesOptions{
		ListOptions: *opts,
	}
	for _, option := range options {
		option(listOpts)
	}

	var output []*gogitlab.Branch
	client := c.client
	for {
		result, resp, err := client.Branches.ListBranches(pid, listOpts, gogitlab.WithContext(ctx))
		if err != nil {
			return nil, nil, apperrors.New(err)
		}
		output = append(output, result...)
		if resp.NextPage <= 0 || opts.Page == resp.NextPage {
			break
		}
		if maxItems > 0 && int64(len(output)) >= maxItems {
			break
		}
		opts.Page = resp.NextPage
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
