package gitlab

import (
	"context"
	"net/http"

	gogitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ListPullRequestOption func(*gogitlab.ListProjectMergeRequestsOptions)

func (c *Client) ListPullRequest(
	ctx context.Context,
	pid any,
	paging *basedto.Paging,
	options ...ListPullRequestOption,
) ([]*gogitlab.BasicMergeRequest, *basedto.PagingMeta, error) {
	opts, maxItems := createListOpts(paging)
	if maxItems > 0 && maxItems > MaxListPageSize {
		return c.ListAllPullRequests(ctx, pid, paging, options...)
	}

	listOpts := &gogitlab.ListProjectMergeRequestsOptions{
		ListOptions: *opts,
	}
	for _, option := range options {
		option(listOpts)
	}

	output, resp, err := c.client.MergeRequests.ListProjectMergeRequests(pid, listOpts, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}
	return output, &basedto.PagingMeta{
		Offset: int(opts.Page * opts.PerPage),
		Limit:  int(opts.PerPage),
		Total:  int(resp.TotalItems),
	}, nil
}

func (c *Client) ListAllPullRequests(
	ctx context.Context,
	pid any,
	paging *basedto.Paging,
	options ...ListPullRequestOption,
) ([]*gogitlab.BasicMergeRequest, *basedto.PagingMeta, error) {
	opts, maxItems := createListOpts(paging)
	listOpts := &gogitlab.ListProjectMergeRequestsOptions{
		ListOptions: *opts,
	}
	for _, option := range options {
		option(listOpts)
	}

	var output []*gogitlab.BasicMergeRequest
	client := c.client
	for {
		result, resp, err := client.MergeRequests.ListProjectMergeRequests(pid, listOpts, gogitlab.WithContext(ctx))
		if err != nil {
			return nil, nil, apperrors.Wrap(err)
		}
		output = append(output, result...)
		if resp.NextPage <= 0 || listOpts.Page == resp.NextPage {
			break
		}
		if maxItems > 0 && int64(len(output)) >= maxItems {
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
	pid any,
	number int,
) (*gogitlab.MergeRequest, error) {
	output, resp, err := c.client.MergeRequests.GetMergeRequest(pid, int64(number), nil, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, apperrors.Wrap(apperrors.ErrPullRequestNotFound).WithParam("PullRequest", number)
	}
	return output, nil
}

func (c *Client) CreatePullRequestComment(
	ctx context.Context,
	pid any,
	number int,
	body string,
) (*gogitlab.Note, error) {
	note, _, err := c.client.Notes.CreateMergeRequestNote(pid, int64(number),
		&gogitlab.CreateMergeRequestNoteOptions{Body: &body}, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return note, nil
}
