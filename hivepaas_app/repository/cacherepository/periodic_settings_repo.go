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
	periodicSettingsKey = "setting:periodic:all"
)

type PeriodicSettingsRepo interface {
	Get(ctx context.Context) (*cacheentity.PeriodicSettings, error)
	Set(ctx context.Context, settings *cacheentity.PeriodicSettings, exp time.Duration) error
	Del(ctx context.Context) error
}

type periodicSettingsRepo struct {
	client rediscache.Client
}

func NewPeriodicSettingsRepo(client rediscache.Client) PeriodicSettingsRepo {
	return &periodicSettingsRepo{client: client}
}

func (repo *periodicSettingsRepo) Get(
	ctx context.Context,
) (*cacheentity.PeriodicSettings, error) {
	resp, err := redishelper.Get[*cacheentity.PeriodicSettings](ctx, repo.client, periodicSettingsKey)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}

func (repo *periodicSettingsRepo) Set(
	ctx context.Context,
	settings *cacheentity.PeriodicSettings,
	exp time.Duration,
) error {
	err := redishelper.Set(ctx, repo.client, periodicSettingsKey, settings, exp)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *periodicSettingsRepo) Del(
	ctx context.Context,
) error {
	err := redishelper.Del(ctx, repo.client, periodicSettingsKey)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
