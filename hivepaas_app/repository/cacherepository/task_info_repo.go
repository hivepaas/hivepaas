package cacherepository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/redishelper"
)

type TaskInfoRepo interface {
	Get(ctx context.Context, taskID string) (*cacheentity.TaskInfo, error)
	MGet(ctx context.Context, taskIDs []string) (map[string]*cacheentity.TaskInfo, error)
	GetAll(ctx context.Context) (map[string]*cacheentity.TaskInfo, error)
	Set(ctx context.Context, taskID string, taskInfo *cacheentity.TaskInfo, exp time.Duration) error
	Update(ctx context.Context, taskID string, taskInfo *cacheentity.TaskInfo) error
	Del(ctx context.Context, taskID string) error
}

type taskInfoRepo struct {
	client rediscache.Client
}

func NewTaskInfoRepo(client rediscache.Client) TaskInfoRepo {
	return &taskInfoRepo{client: client}
}

func (repo *taskInfoRepo) Get(
	ctx context.Context,
	taskID string,
) (*cacheentity.TaskInfo, error) {
	resp, err := redishelper.Get[*cacheentity.TaskInfo](ctx, repo.client, repo.formatKey(taskID))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}

func (repo *taskInfoRepo) MGet(
	ctx context.Context,
	taskIDs []string,
) (map[string]*cacheentity.TaskInfo, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	keys := make([]string, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		keys = append(keys, repo.formatKey(taskID))
	}
	return repo.mGet(ctx, keys)
}

func (repo *taskInfoRepo) GetAll(
	ctx context.Context,
) (map[string]*cacheentity.TaskInfo, error) {
	keys, err := redishelper.ScanKeys(ctx, repo.client, repo.formatKey("*"))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if len(keys) == 0 {
		return nil, nil
	}
	return repo.mGet(ctx, keys)
}

func (repo *taskInfoRepo) mGet(
	ctx context.Context,
	keys []string,
) (map[string]*cacheentity.TaskInfo, error) {
	resp, err := redishelper.MGet[*cacheentity.TaskInfo](ctx, repo.client, keys...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	result := make(map[string]*cacheentity.TaskInfo, len(resp))
	for _, info := range resp {
		if info != nil {
			result[info.ID] = info
		}
	}
	return result, nil
}

func (repo *taskInfoRepo) Set(
	ctx context.Context,
	taskID string,
	taskInfo *cacheentity.TaskInfo,
	exp time.Duration,
) error {
	err := redishelper.Set(ctx, repo.client, repo.formatKey(taskID), taskInfo, exp)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *taskInfoRepo) Update(
	ctx context.Context,
	taskID string,
	taskInfo *cacheentity.TaskInfo,
) error {
	err := redishelper.SetXX(ctx, repo.client, repo.formatKey(taskID), taskInfo, redis.KeepTTL)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *taskInfoRepo) Del(ctx context.Context, taskID string) error {
	err := redishelper.Del(ctx, repo.client, repo.formatKey(taskID))
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *taskInfoRepo) formatKey(taskID string) string {
	return fmt.Sprintf("task:%s:info", taskID)
}
