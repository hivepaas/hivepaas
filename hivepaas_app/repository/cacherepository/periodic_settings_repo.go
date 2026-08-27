package cacherepository

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/redishelper"
)

const (
	periodicScheduleKey = "queue:periodic:schedule"
)

type PeriodicSettingsRepo interface {
	GetDueJobIDs(ctx context.Context, nowSecs int64, limit int64) ([]string, error)
	ScheduleJob(ctx context.Context, jobID string, nextRunSecs int64) error
	RemoveJob(ctx context.Context, jobID string) error
	ResetSchedule(ctx context.Context) error
}

type periodicSettingsRepo struct {
	client rediscache.Client
}

func NewPeriodicSettingsRepo(
	client rediscache.Client,
) PeriodicSettingsRepo {
	return &periodicSettingsRepo{
		client: client,
	}
}

func (repo *periodicSettingsRepo) GetDueJobIDs(
	ctx context.Context,
	nowSecs int64,
	limit int64,
) ([]string, error) {
	members, err := redishelper.ZRangeByScore(ctx, repo.client, periodicScheduleKey, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    strconv.FormatInt(nowSecs, 10),
		Offset: 0,
		Count:  limit,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return members, nil
}

func (repo *periodicSettingsRepo) ScheduleJob(
	ctx context.Context,
	jobID string,
	nextRunSecs int64,
) error {
	err := redishelper.ZAdd(ctx, repo.client, periodicScheduleKey, redis.Z{
		Score:  float64(nextRunSecs),
		Member: jobID,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *periodicSettingsRepo) RemoveJob(
	ctx context.Context,
	jobID string,
) error {
	err := redishelper.ZRem(ctx, repo.client, periodicScheduleKey, jobID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *periodicSettingsRepo) ResetSchedule(
	ctx context.Context,
) error {
	err := redishelper.Del(ctx, repo.client, periodicScheduleKey)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
