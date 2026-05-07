package applog

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/redishelper"
)

const (
	storeExpiration = 4 * time.Hour
)

type Store struct {
	redisClient       redis.UniversalClient
	Key               string
	storeLocal        bool
	storeRemote       bool
	remoteInitialized atomic.Bool
	mu                sync.RWMutex
	frames            []*LogFrame
}

func (s *Store) Add(ctx context.Context, frames ...*LogFrame) error {
	if s.storeLocal {
		s.mu.Lock()
		s.frames = append(s.frames, frames...)
		s.mu.Unlock()
	}

	if s.storeRemote {
		// Store log data in redis
		err := redishelper.RPush(ctx, s.redisClient, s.Key, frames...)
		if err != nil {
			return apperrors.New(err).WithMsgLog("failed to push log frames to redis")
		}

		if s.remoteInitialized.CompareAndSwap(false, true) {
			s.redisClient.Expire(ctx, s.Key, storeExpiration)
		}

		// Notify consumers about the new data
		_, err = s.redisClient.Publish(ctx, s.Key, buildMessage(CommandNewData)).Result()
		if err != nil {
			return apperrors.New(err).WithMsgLog("failed to notify consumers about the new data")
		}
	}

	return nil
}

func (s *Store) GetData(ctx context.Context, fromIndex int64) ([]*LogFrame, error) {
	if s.storeLocal {
		return s.GetLocalData(ctx, fromIndex)
	}
	if s.storeRemote {
		return s.GetRemoteData(ctx, fromIndex)
	}
	return nil, apperrors.NewUnavailable("Log store")
}

func (s *Store) GetLocalData(ctx context.Context, fromIndex int64) ([]*LogFrame, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if fromIndex >= int64(len(s.frames)) {
		return nil, nil
	}
	return s.frames[fromIndex:], nil
}

func (s *Store) GetRemoteData(ctx context.Context, fromIndex int64) ([]*LogFrame, error) {
	frames, err := redishelper.LRange[*LogFrame](ctx, s.redisClient, s.Key, fromIndex, -1)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return frames, nil
}

func (s *Store) Close() (err error) {
	if s.storeRemote {
		ctx := context.Background()
		// Send close-msg to consumers
		_, e := s.redisClient.Publish(ctx, s.Key, buildMessage(CommandClosed)).Result()
		if e != nil {
			err = errors.Join(err, apperrors.New(err).WithMsgLog("failed to notify consumers"))
		}
		// Delete log data in redis
		e = redishelper.Del(ctx, s.redisClient, s.Key)
		if e != nil {
			err = errors.Join(err, apperrors.New(err).WithMsgLog("failed to remove data from redis"))
		}
	}

	return err
}

func newStore(
	key string,
	storeLocal bool,
	storeRemote bool,
	redisClient redis.UniversalClient,
) *Store {
	s := &Store{
		redisClient: redisClient,
		Key:         key,
		storeLocal:  storeLocal,
		storeRemote: storeRemote,
	}
	if storeLocal {
		s.frames = make([]*LogFrame, 0, 100) //nolint:mnd
	}
	return s
}

func NewRemoteStore(
	key string,
	storeLocal bool,
	redisClient redis.UniversalClient,
) *Store {
	return newStore(key, storeLocal, true, redisClient)
}

func NewLocalStore(
	key string,
) *Store {
	return newStore(key, true, false, nil)
}
