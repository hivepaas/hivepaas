package gitlab

import (
	"context"

	gogitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ListProjectOption func(*gogitlab.ListProjectsOptions)

func (c *Client) ListProjects(
	ctx context.Context,
	paging *basedto.Paging,
	options ...ListProjectOption,
) ([]*gogitlab.Project, *basedto.PagingMeta, error) {
	opts, maxItems := createListOpts(paging)
	if maxItems > 0 && maxItems > MaxListPageSize {
		return c.ListAllProjects(ctx, paging, options...)
	}

	listOpts := &gogitlab.ListProjectsOptions{
		ListOptions: *opts,
		Membership:  new(true),
	}
	for _, option := range options {
		option(listOpts)
	}

	output, resp, err := c.client.Projects.ListProjects(listOpts)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}
	return output, &basedto.PagingMeta{
		Offset: int(opts.Page * opts.PerPage),
		Limit:  int(opts.PerPage),
		Total:  int(resp.TotalItems),
	}, nil
}

func (c *Client) ListAllProjects(
	ctx context.Context,
	paging *basedto.Paging,
	options ...ListProjectOption,
) ([]*gogitlab.Project, *basedto.PagingMeta, error) {
	opts, maxItems := createListOpts(paging)
	projOpts := &gogitlab.ListProjectsOptions{
		Membership:  new(true),
		ListOptions: *opts,
	}
	for _, option := range options {
		option(projOpts)
	}

	var output []*gogitlab.Project
	client := c.client
	for {
		result, resp, err := client.Projects.ListProjects(projOpts)
		if err != nil {
			return nil, nil, apperrors.Wrap(err)
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
