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

// DueJob is a job whose next run has come, and when it was due.
type DueJob struct {
	ID      string
	DueSecs int64
}

type PeriodicSettingsRepo interface {
	// GetDueJobs is the jobs due by nowSecs, the earliest first, limit at most.
	GetDueJobs(ctx context.Context, nowSecs int64, limit int64) ([]DueJob, error)
	// ScheduleJobs sets when each job next runs, all in one call. With
	// keepExisting, a job already scheduled keeps its time: only one not in the
	// schedule is added.
	ScheduleJobs(ctx context.Context, nextRunSecs map[string]int64, keepExisting bool) error
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

func (repo *periodicSettingsRepo) GetDueJobs(
	ctx context.Context,
	nowSecs int64,
	limit int64,
) ([]DueJob, error) {
	members, err := redishelper.ZRangeByScoreWithScores(ctx, repo.client, periodicScheduleKey, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    strconv.FormatInt(nowSecs, 10),
		Offset: 0,
		Count:  limit,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	due := make([]DueJob, 0, len(members))
	for _, member := range members {
		id, _ := member.Member.(string)
		due = append(due, DueJob{ID: id, DueSecs: int64(member.Score)})
	}
	return due, nil
}

func (repo *periodicSettingsRepo) ScheduleJobs(
	ctx context.Context,
	nextRunSecs map[string]int64,
	keepExisting bool,
) error {
	members := make([]redis.Z, 0, len(nextRunSecs))
	for id, at := range nextRunSecs {
		members = append(members, redis.Z{Score: float64(at), Member: id})
	}
	add := redishelper.ZAdd
	if keepExisting {
		add = redishelper.ZAddNX
	}
	if err := add(ctx, repo.client, periodicScheduleKey, members...); err != nil {
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
