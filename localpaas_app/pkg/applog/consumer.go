package applog

import (
	"context"
	"errors"
	"sync"

	"github.com/redis/go-redis/v9"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/batchrecvchan"
	"github.com/localpaas/localpaas/localpaas_app/pkg/redishelper"
)

type Consumer struct {
	redisClient redis.UniversalClient
	key         string
}

func NewConsumer(
	key string,
	redisClient redis.UniversalClient,
) *Consumer {
	return &Consumer{
		key:         key,
		redisClient: redisClient,
	}
}

func (c *Consumer) StartConsuming(
	ctx context.Context,
	options batchrecvchan.Options,
) (<-chan []*LogFrame, func() error, error) {
	batchChan := batchrecvchan.NewChan[*LogFrame](options)
	pubSub := c.redisClient.Subscribe(ctx, c.key)
	var closeOnce sync.Once
	var closeErr error

	closeAllFunc := func() error { //nolint:contextcheck
		closeOnce.Do(func() {
			closeErr = errors.Join(
				pubSub.Unsubscribe(context.Background()), // `ctx` may expire at this point
				pubSub.Close(),
				batchChan.Close(),
			)
		})
		return closeErr
	}

	go func() {
		defer func() {
			_ = recover()
			_ = closeAllFunc()
		}()

		frameIndex := int64(0)

		// Query all current logs
		frames, _ := c.getData(ctx, &frameIndex)
		batchChan.Send(frames...)

		for msg := range pubSub.Channel() {
			cmd, _ := parseMessage(msg.Payload)
			switch cmd {
			case CommandNewData:
				frames, err := c.getData(ctx, &frameIndex)
				if err != nil {
					batchChan.Send(&LogFrame{
						Type: LogTypeErr,
						Data: "failed to get log data: " + err.Error(),
					})
					continue
				}
				batchChan.Send(frames...)

			case CommandClosed:
				return
			}
		}
	}()

	return batchChan.Receiver(), closeAllFunc, nil
}

func (c *Consumer) getData(
	ctx context.Context,
	frameIndex *int64,
) (frames []*LogFrame, err error) {
	frames, err = redishelper.LRange[*LogFrame](ctx, c.redisClient, c.key, *frameIndex, -1)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	*frameIndex += int64(len(frames))
	return frames, nil
}

func (c *Consumer) GetAllData(
	ctx context.Context,
) ([]*LogFrame, error) {
	frameIndex := int64(0)
	frames, err := c.getData(ctx, &frameIndex)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return frames, nil
}
