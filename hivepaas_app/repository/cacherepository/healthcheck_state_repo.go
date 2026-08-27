package cacherepository

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/redishelper"
)

const (
	healthCheckStateMapKey = "healthcheck:state:map"
)

type HealthcheckStateRepo interface {
	Get(ctx context.Context, id string) (*cacheentity.HealthcheckState, error)
	GetAll(ctx context.Context) (map[string]*cacheentity.HealthcheckState, error)
	Set(ctx context.Context, id string, notifEvent *cacheentity.HealthcheckState, exp time.Duration) error
	Del(ctx context.Context, id string) error
}

type healthcheckStateRepo struct {
	client rediscache.Client
}

func NewHealthcheckStateRepo(client rediscache.Client) HealthcheckStateRepo {
	return &healthcheckStateRepo{client: client}
}

func (repo *healthcheckStateRepo) Get(
	ctx context.Context,
	id string,
) (*cacheentity.HealthcheckState, error) {
	resp, err := redishelper.HGet[*cacheentity.HealthcheckState](ctx, repo.client, healthCheckStateMapKey, id)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}

func (repo *healthcheckStateRepo) GetAll(
	ctx context.Context,
) (map[string]*cacheentity.HealthcheckState, error) {
	resp, err := redishelper.HGetAll[*cacheentity.HealthcheckState](ctx, repo.client, healthCheckStateMapKey)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}

func (repo *healthcheckStateRepo) Set(
	ctx context.Context,
	id string,
	notifEvent *cacheentity.HealthcheckState,
	exp time.Duration,
) error {
	err := redishelper.HSet(ctx, repo.client, healthCheckStateMapKey, id, notifEvent, exp)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *healthcheckStateRepo) Del(
	ctx context.Context,
	id string,
) error {
	err := redishelper.Del(ctx, repo.client, id)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
