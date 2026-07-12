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

type DeploymentInfoRepo interface {
	Get(ctx context.Context, deploymentID string) (*cacheentity.DeploymentInfo, error)
	MGet(ctx context.Context, deploymentIDs []string) (map[string]*cacheentity.DeploymentInfo, error)
	GetAllOfApp(ctx context.Context, appID string) (map[string]*cacheentity.DeploymentInfo, error)
	GetAll(ctx context.Context) (map[string]*cacheentity.DeploymentInfo, error)
	Set(ctx context.Context, deploymentID string, deploymentInfo *cacheentity.DeploymentInfo, exp time.Duration) error
	Del(ctx context.Context, deploymentID string) error
}

type deploymentInfoRepo struct {
	client rediscache.Client
}

func NewDeploymentInfoRepo(client rediscache.Client) DeploymentInfoRepo {
	return &deploymentInfoRepo{client: client}
}

func (repo *deploymentInfoRepo) Get(
	ctx context.Context,
	deploymentID string,
) (*cacheentity.DeploymentInfo, error) {
	resp, err := redishelper.Get[*cacheentity.DeploymentInfo](ctx, repo.client, repo.formatKey(deploymentID))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}

func (repo *deploymentInfoRepo) MGet(
	ctx context.Context,
	deploymentIDs []string,
) (map[string]*cacheentity.DeploymentInfo, error) {
	if len(deploymentIDs) == 0 {
		return nil, nil
	}
	keys := make([]string, 0, len(deploymentIDs))
	for _, deploymentID := range deploymentIDs {
		keys = append(keys, repo.formatKey(deploymentID))
	}
	return repo.mGet(ctx, keys)
}

func (repo *deploymentInfoRepo) GetAllOfApp(
	ctx context.Context,
	appID string,
) (map[string]*cacheentity.DeploymentInfo, error) {
	keys, err := redishelper.ScanKeys(ctx, repo.client, repo.formatKey("*"))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if len(keys) == 0 {
		return nil, nil
	}
	deployments, err := repo.mGet(ctx, keys)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	appDeployments := make(map[string]*cacheentity.DeploymentInfo, len(deployments))
	for k, deployment := range deployments {
		if deployment.AppID != appID {
			continue
		}
		appDeployments[k] = deployment
	}
	return appDeployments, nil
}

func (repo *deploymentInfoRepo) GetAll(
	ctx context.Context,
) (map[string]*cacheentity.DeploymentInfo, error) {
	keys, err := redishelper.ScanKeys(ctx, repo.client, repo.formatKey("*"))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if len(keys) == 0 {
		return nil, nil
	}
	return repo.mGet(ctx, keys)
}

func (repo *deploymentInfoRepo) mGet(ctx context.Context, keys []string) (
	map[string]*cacheentity.DeploymentInfo, error) {
	resp, err := redishelper.MGet[*cacheentity.DeploymentInfo](ctx, repo.client, keys...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	result := make(map[string]*cacheentity.DeploymentInfo, len(resp))
	for _, info := range resp {
		if info != nil {
			result[info.ID] = info
		}
	}
	return result, nil
}

func (repo *deploymentInfoRepo) Set(
	ctx context.Context,
	deploymentID string,
	deploymentInfo *cacheentity.DeploymentInfo,
	exp time.Duration,
) error {
	err := redishelper.Set(ctx, repo.client, repo.formatKey(deploymentID), deploymentInfo, exp)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *deploymentInfoRepo) Del(ctx context.Context, deploymentID string) error {
	err := redishelper.Del(ctx, repo.client, repo.formatKey(deploymentID))
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *deploymentInfoRepo) formatKey(deploymentID string) string {
	return fmt.Sprintf("deployment:%s:info", deploymentID)
}
