package gocronqueue

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/redishelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
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
	err := c.send(ctx, &Message{
		StartScheduler: true,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (c *Client) StopScheduler(ctx context.Context) error {
	err := c.send(ctx, &Message{
		StopScheduler: true,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (c *Client) ScheduleTask(ctx context.Context, tasks ...*entity.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	err := c.send(ctx, &Message{
		SchedTasks: tasks,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (c *Client) UnscheduleTask(ctx context.Context, taskIDs ...string) error {
	if len(taskIDs) == 0 {
		return nil
	}
	err := c.send(ctx, &Message{
		UnschedTaskIDs: taskIDs,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// send puts a message on the list every server reads, saying when it was sent.
func (c *Client) send(ctx context.Context, msg *Message) error {
	msg.SentAt = timeutil.NowUTC()
	return redishelper.RPush(ctx, c.redisClient, taskQueueCtrlKey, msg) //nolint:wrapcheck // its callers wrap
}
