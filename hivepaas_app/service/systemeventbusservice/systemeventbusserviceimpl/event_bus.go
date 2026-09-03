package systemeventbusserviceimpl

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/redishelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
)

const (
	systemEventsRedisChannel = "system:events"
	systemEventsBufferSize   = 32
)

func (s *service) Publish(
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
		return hperrors.Wrap(err)
	}
	err = redishelper.Publish(ctx, s.client, systemEventsRedisChannel, string(bytes))
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *service) Subscribe(
	eventTypes ...base.SystemEventType,
) (<-chan *entity.SystemEvent, func()) {
	ch := make(chan *entity.SystemEvent, systemEventsBufferSize)
	s.mu.Lock()
	for _, t := range eventTypes {
		s.listeners[t] = append(s.listeners[t], ch)
	}
	s.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			for _, t := range eventTypes {
				s.listeners[t] = gofn.Filter(s.listeners[t], func(c chan *entity.SystemEvent) bool {
					return c != ch
				})
			}
			close(ch)
		})
	}
	return ch, unsubscribe
}

func (s *service) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.pubSub != nil {
		s.mu.Unlock()
		return nil
	}
	subCtx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel
	s.pubSub = s.client.Subscribe(subCtx, systemEventsRedisChannel)
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer safego.RecoverWithLogger(s.logger, "systemEventBus.subscribeLoop")

		ch := s.pubSub.Channel()
		for msg := range ch {
			var event entity.SystemEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				s.logger.Warnf("failed to unmarshal system event: %v", err)
				continue
			}

			s.handleEvent(subCtx, &event)

			s.mu.RLock()
			targets := s.listeners[event.Type]
			for _, targetCh := range targets {
				select {
				case targetCh <- &event:
				default:
					s.logger.Warnf("system event channel buffer full for type %s, dropping event", event.Type)
				}
			}
			s.mu.RUnlock()
		}
	}()

	return nil
}

func (s *service) Stop() error {
	s.mu.Lock()
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}
	pubSub := s.pubSub
	s.pubSub = nil
	s.mu.Unlock()

	var err error
	if pubSub != nil {
		err = errors.Join(
			pubSub.Unsubscribe(context.Background()),
			pubSub.Close(),
		)
	}
	s.wg.Wait()
	return err
}
