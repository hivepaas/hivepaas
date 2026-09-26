package cacherepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
)

// MCPPlanRepo keeps the plans an MCP tool made until one is applied or expires.
// A plan arrives sealed: this repository stores bytes it cannot read.
type MCPPlanRepo interface {
	Set(ctx context.Context, planID string, sealed []byte, exp time.Duration) error
	// GetDel takes a plan, so that it can be applied once. Nil when there is none.
	GetDel(ctx context.Context, planID string) ([]byte, error)
}

type mcpPlanRepo struct {
	client rediscache.Client
}

func NewMCPPlanRepo(client rediscache.Client) MCPPlanRepo {
	return &mcpPlanRepo{client: client}
}

func (repo *mcpPlanRepo) Set(ctx context.Context, planID string, sealed []byte, exp time.Duration) error {
	if err := repo.client.Set(ctx, repo.formatKey(planID), sealed, exp).Err(); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *mcpPlanRepo) GetDel(ctx context.Context, planID string) ([]byte, error) {
	sealed, err := repo.client.GetDel(ctx, repo.formatKey(planID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return sealed, nil
}

func (repo *mcpPlanRepo) formatKey(planID string) string {
	return fmt.Sprintf("mcp:plan:%s", planID)
}
