package cacherepository

import (
	"context"
	"fmt"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/redishelper"
)

type AppSecretAttemptRepo interface {
	Get(ctx context.Context, userID string) (*cacheentity.AppSecretAttempt, error)
	Set(ctx context.Context, userID string, attempt *cacheentity.AppSecretAttempt, exp time.Duration) error
	Del(ctx context.Context, userID string) error
}

type appSecretAttemptRepo struct {
	client rediscache.Client
}

func NewAppSecretAttemptRepo(client rediscache.Client) AppSecretAttemptRepo {
	return &appSecretAttemptRepo{client: client}
}

func (repo *appSecretAttemptRepo) Get(
	ctx context.Context,
	userID string,
) (*cacheentity.AppSecretAttempt, error) {
	resp, err := redishelper.Get[*cacheentity.AppSecretAttempt](ctx, repo.client, repo.formatKey(userID))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}

func (repo *appSecretAttemptRepo) Set(
	ctx context.Context,
	userID string,
	attempt *cacheentity.AppSecretAttempt,
	exp time.Duration,
) error {
	err := redishelper.Set(ctx, repo.client, repo.formatKey(userID), attempt, exp)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *appSecretAttemptRepo) Del(ctx context.Context, userID string) error {
	err := redishelper.Del(ctx, repo.client, repo.formatKey(userID))
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *appSecretAttemptRepo) formatKey(userID string) string {
	return fmt.Sprintf("app-secret-attempt:%s", userID)
}
