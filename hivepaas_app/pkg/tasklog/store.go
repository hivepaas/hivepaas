package tasklog

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/redact"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/redishelper"
)

const (
	KB              = 1024
	DefaultMaxSize  = 64 * KB
	storeExpiration = 4 * time.Hour
)

type Store struct {
	redisClient       redis.UniversalClient
	Key               string
	storeLocal        bool
	storeRemote       bool
	remoteInitialized atomic.Bool
	redactor          *redact.Redactor
	mu                sync.RWMutex
	frames            []*LogFrame
	totalSize         int64
	maxSize           int64
	onFlush           func(ctx context.Context, frames []*LogFrame) error
}

func (s *Store) GetRedactor() *redact.Redactor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.redactor
}

func (s *Store) SetRedactor(redactor *redact.Redactor) {
	s.mu.Lock()
	s.redactor = redactor
	s.mu.Unlock()
}

func (s *Store) UpdateRedactorAddSecrets(secrets []string) {
	s.mu.Lock()
	if s.redactor == nil {
		s.redactor = redact.New(secrets)
	} else {
		s.redactor.AddSecrets(secrets)
	}
	s.mu.Unlock()
}

func (s *Store) SetOnFlush(maxSize int64, onFlush func(context.Context, []*LogFrame) error) {
	s.mu.Lock()
	s.maxSize = maxSize
	s.onFlush = onFlush
	if s.frames == nil {
		s.frames = make([]*LogFrame, 0, 100) //nolint:mnd
	}
	s.mu.Unlock()
}

func (s *Store) Add(ctx context.Context, frames ...*LogFrame) error {
	var framesSize int64
	for _, f := range frames {
		framesSize += int64(len(f.Data))
	}

	s.mu.Lock()
	if s.storeLocal || s.onFlush != nil {
		s.frames = append(s.frames, frames...)
	}
	s.totalSize += framesSize
	s.mu.Unlock()

	if s.storeRemote {
		// Store log data in redis
		err := redishelper.RPush(ctx, s.redisClient, s.Key, frames...)
		if err != nil {
			return apperrors.Wrap(err).WithMsgLog("failed to push log frames to redis")
		}

		if s.remoteInitialized.CompareAndSwap(false, true) {
			s.redisClient.Expire(ctx, s.Key, storeExpiration)
		}

		// Notify consumers about the new data
		_, err = s.redisClient.Publish(ctx, s.Key, buildMessage(CommandNewData)).Result()
		if err != nil {
			return apperrors.Wrap(err).WithMsgLog("failed to notify consumers about the new data")
		}
	}

	s.mu.Lock()
	shouldFlush := s.onFlush != nil && s.maxSize > 0 && s.totalSize >= s.maxSize
	var framesToFlush []*LogFrame
	if shouldFlush {
		framesToFlush = s.frames
		s.frames = make([]*LogFrame, 0, 100) //nolint:mnd
		s.totalSize = 0
	}
	s.mu.Unlock()

	if shouldFlush && len(framesToFlush) > 0 {
		if s.storeRemote {
			_ = redishelper.Del(ctx, s.redisClient, s.Key)
		}
		if err := s.onFlush(ctx, framesToFlush); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) AddRedacted(ctx context.Context, frames ...*LogFrame) error {
	s.mu.RLock()
	r := s.redactor
	s.mu.RUnlock()

	if r == nil {
		return s.Add(ctx, frames...)
	}

	input := make([]string, len(frames))
	for i, frame := range frames {
		input[i] = frame.Data
	}

	input = r.Slice(input)
	for i, frame := range frames {
		frame.Data = input[i]
	}
	return s.Add(ctx, frames...)
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

func (s *Store) Reset() (err error) {
	if s.storeRemote {
		ctx := context.Background()
		// Send close-msg to consumers
		_, e := s.redisClient.Publish(ctx, s.Key, buildMessage(CommandClosed)).Result()
		if e != nil {
			err = errors.Join(err, apperrors.Wrap(err).WithMsgLog("failed to notify consumers"))
		}
		// Delete log data in redis
		e = redishelper.Del(ctx, s.redisClient, s.Key)
		if e != nil {
			err = errors.Join(err, apperrors.Wrap(err).WithMsgLog("failed to remove data from redis"))
		}
	}
	if s.storeLocal {
		s.mu.Lock()
		s.frames = make([]*LogFrame, 0, 100) //nolint:mnd
		s.totalSize = 0
		s.mu.Unlock()
	}
	return err
}

func (s *Store) Close() (err error) {
	return s.Reset()
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
	redisClient redis.UniversalClient,
) *Store {
	return newStore(key, true, true, redisClient)
}

func NewLocalStore(
	key string,
) *Store {
	return newStore(key, true, false, nil)
}

func NewNullStore() *Store {
	return &Store{}
}
