package cacherepository

import (
	"context"
	"fmt"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/redishelper"
)

const (
	githubAppManifestKey = "github:app:%s:manifest"
)

type GithubAppManifestRepo interface {
	Get(ctx context.Context, settingID string) (*cacheentity.GithubAppManifest, error)
	Set(ctx context.Context, settingID string, manifest *cacheentity.GithubAppManifest, exp time.Duration) error
	Del(ctx context.Context, settingID string) error
}

type githubAppManifestRepo struct {
	client rediscache.Client
}

func NewGithubAppManifestRepo(client rediscache.Client) GithubAppManifestRepo {
	return &githubAppManifestRepo{client: client}
}

func (repo *githubAppManifestRepo) Get(
	ctx context.Context,
	settingID string,
) (*cacheentity.GithubAppManifest, error) {
	resp, err := redishelper.Get[*cacheentity.GithubAppManifest](ctx, repo.client, repo.formatKey(settingID))
	if err != nil {
		return nil, apperrors.New(err)
	}
	return resp, nil
}

func (repo *githubAppManifestRepo) Set(
	ctx context.Context,
	settingID string,
	manifest *cacheentity.GithubAppManifest,
	exp time.Duration,
) error {
	err := redishelper.Set(ctx, repo.client, repo.formatKey(settingID), manifest, exp)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *githubAppManifestRepo) Del(
	ctx context.Context,
	settingID string,
) error {
	err := redishelper.Del(ctx, repo.client, repo.formatKey(settingID))
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *githubAppManifestRepo) formatKey(settingID string) string {
	return fmt.Sprintf(githubAppManifestKey, settingID)
}
