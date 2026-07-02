package cacherepository

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/redishelper"
)

const (
	healthCheckSettingsKey = "setting:healthcheck:all"
)

type HealthcheckSettingsRepo interface {
	Get(ctx context.Context) (*cacheentity.HealthcheckSettings, error)
	Set(ctx context.Context, settings *cacheentity.HealthcheckSettings, exp time.Duration) error
	Del(ctx context.Context) error
}

type healthcheckSettingsRepo struct {
	client rediscache.Client
}

func NewHealthcheckSettingsRepo(client rediscache.Client) HealthcheckSettingsRepo {
	return &healthcheckSettingsRepo{client: client}
}

func (repo *healthcheckSettingsRepo) Get(
	ctx context.Context,
) (*cacheentity.HealthcheckSettings, error) {
	resp, err := redishelper.Get[*cacheentity.HealthcheckSettings](ctx, repo.client, healthCheckSettingsKey)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return resp, nil
}

func (repo *healthcheckSettingsRepo) Set(
	ctx context.Context,
	settings *cacheentity.HealthcheckSettings,
	exp time.Duration,
) error {
	err := redishelper.Set(ctx, repo.client, healthCheckSettingsKey, settings, exp)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *healthcheckSettingsRepo) Del(
	ctx context.Context,
) error {
	err := redishelper.Del(ctx, repo.client, healthCheckSettingsKey)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
