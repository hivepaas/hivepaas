package gocronqueue

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/redishelper"
)

type Client struct {
	redisClient redis.UniversalClient
	logger      logging.Logger
}

func NewClient(
	redisClient redis.UniversalClient,
	logger logging.Logger,
) (*Client, error) {
	return &Client{
		redisClient: redisClient,
		logger:      logger,
	}, nil
}

func (c *Client) Close() error {
	// Use shared client, so we don't close it
	return nil
}

func (c *Client) StartScheduler(ctx context.Context) error {
	err := redishelper.RPush(ctx, c.redisClient, taskQueueCtrlKey, &Message{
		StartScheduler: true,
	})
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (c *Client) StopScheduler(ctx context.Context) error {
	err := redishelper.RPush(ctx, c.redisClient, taskQueueCtrlKey, &Message{
		StopScheduler: true,
	})
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (c *Client) ScheduleTask(ctx context.Context, tasks ...*entity.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	err := redishelper.RPush(ctx, c.redisClient, taskQueueCtrlKey, &Message{
		SchedTasks: tasks,
	})
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (c *Client) UnscheduleTask(ctx context.Context, taskIDs ...string) error {
	if len(taskIDs) == 0 {
		return nil
	}
	err := redishelper.RPush(ctx, c.redisClient, taskQueueCtrlKey, &Message{
		UnschedTaskIDs: taskIDs,
	})
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
