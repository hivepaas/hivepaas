package gitea

import (
	"context"

	gogitea "code.gitea.io/sdk/gitea"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ListReposOption func(options *gogitea.ListReposOptions)

func (c *Client) ListRepos(
	ctx context.Context,
	paging *basedto.Paging,
	options ...ListReposOption,
) ([]*gogitea.Repository, *basedto.PagingMeta, error) {
	opts, maxItems := createListOpts(paging)
	if maxItems > 0 && maxItems > MaxListPageSize {
		return c.ListAllRepos(ctx, paging, options...)
	}

	listOpts := gogitea.ListReposOptions{
		ListOptions: *opts,
	}
	for _, option := range options {
		option(&listOpts)
	}

	output, resp, err := c.client.ListMyRepos(listOpts)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}
	return output, &basedto.PagingMeta{
		Offset: opts.Page * opts.PageSize,
		Limit:  opts.PageSize,
		Total:  resp.LastPage * opts.PageSize,
	}, nil
}

func (c *Client) ListAllRepos(
	ctx context.Context,
	paging *basedto.Paging,
	options ...ListReposOption,
) ([]*gogitea.Repository, *basedto.PagingMeta, error) {
	opts, maxItems := createListOpts(paging)
	listOpts := gogitea.ListReposOptions{
		ListOptions: *opts,
	}
	for _, option := range options {
		option(&listOpts)
	}

	var output []*gogitea.Repository
	client := c.client
	for {
		result, resp, err := client.ListMyRepos(listOpts)
		if err != nil {
			return nil, nil, apperrors.Wrap(err)
		}
		output = append(output, result...)
		if resp.NextPage <= 0 || opts.Page == resp.NextPage {
			break
		}
		if maxItems > 0 && len(output) >= maxItems {
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
