package github

import (
	"context"

	gogithub "github.com/google/go-github/v85/github"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

func (c *Client) ListAppRepos(
	ctx context.Context,
	paging *basedto.Paging,
) ([]*gogithub.Repository, *basedto.PagingMeta, error) {
	if !c.IsAppClient() {
		return nil, nil, apperrors.New(ErrGithubAppClientRequired)
	}

	opts, maxItems := createListOpts(paging)
	if maxItems > 0 && maxItems > MaxListPageSize {
		return c.ListAllAppRepos(ctx, paging)
	}

	output, _, err := c.client.Apps.ListRepos(ctx, opts)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}
	return output.Repositories, &basedto.PagingMeta{
		Offset: opts.Page * opts.PerPage,
		Limit:  opts.PerPage,
		Total:  -1,
	}, nil
}

func (c *Client) ListAllAppRepos(
	ctx context.Context,
	paging *basedto.Paging,
) ([]*gogithub.Repository, *basedto.PagingMeta, error) {
	if !c.IsAppClient() {
		return nil, nil, apperrors.New(ErrGithubAppClientRequired)
	}

	opts, maxItems := createListOpts(paging)
	var output []*gogithub.Repository
	client := c.client
	for {
		result, resp, err := client.Apps.ListRepos(ctx, opts)
		if err != nil {
			return nil, nil, apperrors.New(err)
		}
		output = append(output, result.Repositories...)
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

type ListUserRepoOption func(options *gogithub.RepositoryListByAuthenticatedUserOptions)

func (c *Client) ListUserRepos(
	ctx context.Context,
	paging *basedto.Paging,
	options ...ListUserRepoOption,
) ([]*gogithub.Repository, *basedto.PagingMeta, error) {
	if !c.IsTokenClient() {
		return nil, nil, apperrors.New(ErrGithubTokenClientRequired)
	}

	opts, maxItems := createListOpts(paging)
	if maxItems > 0 && maxItems > MaxListPageSize {
		return c.ListAllUserRepos(ctx, paging)
	}

	listOpts := &gogithub.RepositoryListByAuthenticatedUserOptions{
		ListOptions: *opts,
	}
	for _, option := range options {
		option(listOpts)
	}

	output, _, err := c.client.Repositories.ListByAuthenticatedUser(ctx, listOpts)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}
	return output, &basedto.PagingMeta{
		Offset: opts.Page * opts.PerPage,
		Limit:  opts.PerPage,
		Total:  -1,
	}, nil
}

func (c *Client) ListAllUserRepos(
	ctx context.Context,
	paging *basedto.Paging,
	options ...ListUserRepoOption,
) ([]*gogithub.Repository, *basedto.PagingMeta, error) {
	if !c.IsTokenClient() {
		return nil, nil, apperrors.New(ErrGithubTokenClientRequired)
	}

	opts, maxItems := createListOpts(paging)
	listOpts := &gogithub.RepositoryListByAuthenticatedUserOptions{
		ListOptions: *opts,
	}
	for _, option := range options {
		option(listOpts)
	}

	var output []*gogithub.Repository
	client := c.client
	for {
		result, resp, err := client.Repositories.ListByAuthenticatedUser(ctx, listOpts)
		if err != nil {
			return nil, nil, apperrors.New(err)
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
