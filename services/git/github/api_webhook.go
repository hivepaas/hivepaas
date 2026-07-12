package github

import (
	"context"

	gogithub "github.com/google/go-github/v85/github"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (c *Client) GetAppHookConfig(
	ctx context.Context,
) (*gogithub.HookConfig, error) {
	if !c.IsAppClient() {
		return nil, apperrors.Wrap(ErrGithubAppClientRequired)
	}

	output, _, err := c.appClient.Apps.GetHookConfig(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return output, nil
}

type UpdateAppHookOption func(options *gogithub.HookConfig)

func (c *Client) UpdateAppHookConfig(
	ctx context.Context,
	options ...UpdateAppHookOption,
) error {
	if !c.IsAppClient() {
		return apperrors.Wrap(ErrGithubAppClientRequired)
	}

	opts := &gogithub.HookConfig{}
	for _, opt := range options {
		opt(opts)
	}

	_, _, err := c.appClient.Apps.UpdateHookConfig(ctx, opts)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
