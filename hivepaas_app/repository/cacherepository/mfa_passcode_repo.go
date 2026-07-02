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

type MFAPasscodeRepo interface {
	Get(ctx context.Context, userID string) (*cacheentity.MFAPasscode, error)
	TTL(ctx context.Context, userID string) (time.Duration, error)
	Set(ctx context.Context, userID string, passcode *cacheentity.MFAPasscode, exp time.Duration) error
	IncrAttempts(ctx context.Context, userID string, passcode *cacheentity.MFAPasscode) error
	Del(ctx context.Context, userID string) error
}

type mfaPasscodeRepo struct {
	client rediscache.Client
}

func NewMFAPasscodeRepo(client rediscache.Client) MFAPasscodeRepo {
	return &mfaPasscodeRepo{client: client}
}

func (repo *mfaPasscodeRepo) Get(
	ctx context.Context,
	userID string,
) (*cacheentity.MFAPasscode, error) {
	resp, err := redishelper.Get[*cacheentity.MFAPasscode](ctx, repo.client, repo.formatKey(userID))
	if err != nil {
		return nil, apperrors.New(err)
	}
	return resp, nil
}

func (repo *mfaPasscodeRepo) TTL(
	ctx context.Context,
	userID string,
) (time.Duration, error) {
	d, err := repo.client.TTL(ctx, repo.formatKey(userID)).Result()
	if err != nil {
		return 0, apperrors.New(err)
	}
	return d, nil
}

func (repo *mfaPasscodeRepo) Set(
	ctx context.Context,
	userID string,
	passcode *cacheentity.MFAPasscode,
	exp time.Duration,
) error {
	err := redishelper.Set(ctx, repo.client, repo.formatKey(userID), passcode, exp)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *mfaPasscodeRepo) IncrAttempts(
	ctx context.Context,
	userID string,
	passcode *cacheentity.MFAPasscode,
) error {
	passcode.Attempts++
	// Only set value, not set expiration to keep the current expiration value
	err := redishelper.Set(ctx, repo.client, repo.formatKey(userID), passcode, redis.KeepTTL)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *mfaPasscodeRepo) Del(ctx context.Context, userID string) error {
	err := redishelper.Del(ctx, repo.client, repo.formatKey(userID))
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *mfaPasscodeRepo) formatKey(userID string) string {
	return fmt.Sprintf("mfa-passcode:%s", userID)
}
