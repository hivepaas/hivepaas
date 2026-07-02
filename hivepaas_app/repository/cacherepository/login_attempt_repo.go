package cacherepository

import (
	"context"
	"fmt"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/redishelper"
)

type LoginAttemptRepo interface {
	Get(ctx context.Context, userID string) (*cacheentity.LoginAttempt, error)
	Set(ctx context.Context, userID string, attempt *cacheentity.LoginAttempt, exp time.Duration) error
	Del(ctx context.Context, userID string) error
}

type loginAttemptRepo struct {
	client rediscache.Client
}

func NewLoginAttemptRepo(client rediscache.Client) LoginAttemptRepo {
	return &loginAttemptRepo{client: client}
}

func (repo *loginAttemptRepo) Get(
	ctx context.Context,
	userID string,
) (*cacheentity.LoginAttempt, error) {
	resp, err := redishelper.Get[*cacheentity.LoginAttempt](ctx, repo.client, repo.formatKey(userID))
	if err != nil {
		return nil, apperrors.New(err)
	}
	return resp, nil
}

func (repo *loginAttemptRepo) Set(
	ctx context.Context,
	userID string,
	attempt *cacheentity.LoginAttempt,
	exp time.Duration,
) error {
	err := redishelper.Set(ctx, repo.client, repo.formatKey(userID), attempt, exp)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *loginAttemptRepo) Del(ctx context.Context, userID string) error {
	err := redishelper.Del(ctx, repo.client, repo.formatKey(userID))
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *loginAttemptRepo) formatKey(userID string) string {
	return fmt.Sprintf("login-attempt:%s", userID)
}
