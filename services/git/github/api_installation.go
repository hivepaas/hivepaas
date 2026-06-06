package github

import (
	"context"

	gogithub "github.com/google/go-github/v85/github"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type ListInstallationOption func(options *gogithub.ListOptions)

func (c *Client) ListInstallations(
	ctx context.Context,
	paging *basedto.Paging,
	options ...ListInstallationOption,
) ([]*gogithub.Installation, *basedto.PagingMeta, error) {
	if !c.IsAppClient() {
		return nil, nil, apperrors.Wrap(ErrGithubAppClientRequired)
	}

	opts, maxItems := createListOpts(paging)
	if maxItems > 0 && maxItems > MaxListPageSize {
		return c.ListAllInstallations(ctx, paging, options...)
	}
	for _, opt := range options {
		opt(opts)
	}

	output, _, err := c.client.Apps.ListInstallations(ctx, opts)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}
	return output, &basedto.PagingMeta{
		Offset: opts.Page * opts.PerPage,
		Limit:  opts.PerPage,
		Total:  -1,
	}, nil
}

func (c *Client) ListAllInstallations(
	ctx context.Context,
	paging *basedto.Paging,
	options ...ListInstallationOption,
) ([]*gogithub.Installation, *basedto.PagingMeta, error) {
	if !c.IsAppClient() {
		return nil, nil, apperrors.Wrap(ErrGithubAppClientRequired)
	}

	opts, maxItems := createListOpts(paging)
	for _, opt := range options {
		opt(opts)
	}

	var output []*gogithub.Installation
	client := c.client
	for {
		result, resp, err := client.Apps.ListInstallations(ctx, opts)
		if err != nil {
			return nil, nil, apperrors.Wrap(err)
		}
		output = append(output, result...)
		if resp.NextPage <= 0 || opts.Page == resp.NextPage || resp.Rate.Remaining <= 0 {
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
