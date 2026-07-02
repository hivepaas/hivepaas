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
	healthCheckNotifEventMapKey = "notif:healthcheck:ts"
)

type HealthcheckNotifEventRepo interface {
	GetAll(ctx context.Context) (map[string]*cacheentity.HealthcheckNotifEvent, error)
	Set(ctx context.Context, id string, notifEvent *cacheentity.HealthcheckNotifEvent, exp time.Duration) error
	Del(ctx context.Context, id string) error
}

type healthcheckNotifEventRepo struct {
	client rediscache.Client
}

func NewHealthcheckNotifEventRepo(client rediscache.Client) HealthcheckNotifEventRepo {
	return &healthcheckNotifEventRepo{client: client}
}

func (repo *healthcheckNotifEventRepo) GetAll(
	ctx context.Context,
) (map[string]*cacheentity.HealthcheckNotifEvent, error) {
	resp, err := redishelper.HGetAll[*cacheentity.HealthcheckNotifEvent](ctx, repo.client, healthCheckNotifEventMapKey)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return resp, nil
}

func (repo *healthcheckNotifEventRepo) Set(
	ctx context.Context,
	id string,
	notifEvent *cacheentity.HealthcheckNotifEvent,
	exp time.Duration,
) error {
	err := redishelper.HSet(ctx, repo.client, healthCheckNotifEventMapKey, id, notifEvent, exp)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *healthcheckNotifEventRepo) Del(
	ctx context.Context,
	id string,
) error {
	err := redishelper.Del(ctx, repo.client, id)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
