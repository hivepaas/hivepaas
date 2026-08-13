package cacherepository

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/redishelper"
)

const (
	systemEventsRedisChannel = "system:events"
	systemEventsBufferSize   = 32
)

type SystemEventBus interface {
	Publish(ctx context.Context, eventType base.SystemEventType, payload ...string) error
	Subscribe(eventTypes ...base.SystemEventType) (<-chan *entity.SystemEvent, func())

	Start(ctx context.Context) error
	Stop() error
}

type systemEventBus struct {
	client rediscache.Client
	logger logging.Logger

	mu        sync.RWMutex
	listeners map[base.SystemEventType][]chan *entity.SystemEvent

	pubSub     *redis.PubSub
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}

func NewSystemEventBus(
	client rediscache.Client,
	logger logging.Logger,
) SystemEventBus {
	return &systemEventBus{
		client:    client,
		logger:    logger,
		listeners: make(map[base.SystemEventType][]chan *entity.SystemEvent),
	}
}

func (b *systemEventBus) Publish(
	ctx context.Context,
	eventType base.SystemEventType,
	payload ...string,
) error {
	var p string
	if len(payload) > 0 {
		p = payload[0]
	}
	event := entity.SystemEvent{
		Type:      eventType,
		Payload:   p,
		CreatedAt: time.Now().UTC(),
	}
	bytes, err := json.Marshal(event)
	if err != nil {
		return apperrors.Wrap(err)
	}
	err = redishelper.Publish(ctx, b.client, systemEventsRedisChannel, string(bytes))
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (b *systemEventBus) Subscribe(
	eventTypes ...base.SystemEventType,
) (<-chan *entity.SystemEvent, func()) {
	ch := make(chan *entity.SystemEvent, systemEventsBufferSize)
	b.mu.Lock()
	for _, t := range eventTypes {
		b.listeners[t] = append(b.listeners[t], ch)
	}
	b.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			for _, t := range eventTypes {
				b.listeners[t] = gofn.Filter(b.listeners[t], func(c chan *entity.SystemEvent) bool {
					return c != ch
				})
			}
			close(ch)
		})
	}
	return ch, unsubscribe
}

func (b *systemEventBus) Start(ctx context.Context) error {
	b.mu.Lock()
	if b.pubSub != nil {
		b.mu.Unlock()
		return nil
	}
	subCtx, cancel := context.WithCancel(ctx)
	b.cancelFunc = cancel
	b.pubSub = b.client.Subscribe(subCtx, systemEventsRedisChannel)
	b.mu.Unlock()

	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		defer func() {
			_ = recover()
		}()

		ch := b.pubSub.Channel()
		for msg := range ch {
			var event entity.SystemEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				b.logger.Warnf("failed to unmarshal system event: %v", err)
				continue
			}

			b.mu.RLock()
			targets := b.listeners[event.Type]
			for _, targetCh := range targets {
				select {
				case targetCh <- &event:
				default:
					b.logger.Warnf("system event channel buffer full for type %s, dropping event", event.Type)
				}
			}
			b.mu.RUnlock()
		}
	}()

	return nil
}

func (b *systemEventBus) Stop() error {
	b.mu.Lock()
	if b.cancelFunc != nil {
		b.cancelFunc()
		b.cancelFunc = nil
	}
	pubSub := b.pubSub
	b.pubSub = nil
	b.mu.Unlock()

	var err error
	if pubSub != nil {
		err = errors.Join(
			pubSub.Unsubscribe(context.Background()),
			pubSub.Close(),
		)
	}
	b.wg.Wait()
	return err
}
