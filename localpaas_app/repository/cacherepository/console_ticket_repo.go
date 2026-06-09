package cacherepository

import (
	"context"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/infra/rediscache"
	"github.com/localpaas/localpaas/localpaas_app/pkg/redishelper"
)

type ConsoleTicketRepo interface {
	Get(ctx context.Context, key string) (*cacheentity.ConsoleTicket, error)
	Set(ctx context.Context, key string, ticket *cacheentity.ConsoleTicket, exp time.Duration) error
	Del(ctx context.Context, key string) error
}

type consoleTicketRepo struct {
	client rediscache.Client
}

func NewConsoleTicketRepo(client rediscache.Client) ConsoleTicketRepo {
	return &consoleTicketRepo{client: client}
}

func (repo *consoleTicketRepo) Get(
	ctx context.Context,
	key string,
) (*cacheentity.ConsoleTicket, error) {
	resp, err := redishelper.Get[*cacheentity.ConsoleTicket](ctx, repo.client, key)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}

func (repo *consoleTicketRepo) Set(
	ctx context.Context,
	key string,
	ticket *cacheentity.ConsoleTicket,
	exp time.Duration,
) error {
	err := redishelper.Set(ctx, repo.client, key, ticket, exp)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *consoleTicketRepo) Del(ctx context.Context, key string) error {
	err := redishelper.Del(ctx, repo.client, key)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
