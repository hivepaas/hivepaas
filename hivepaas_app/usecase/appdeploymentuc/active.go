package appdeploymentuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

// activeDeploymentsMax bounds what is read of the app's deployments that have
// not ended: one running, a few queued behind it at most.
const activeDeploymentsMax = 10

// GetActiveDeployment answers whether the app has a deployment queued or
// running, and which: what a screen asks every few seconds, so it reads no more
// than that - a few rows of the app's, and their states from the cache.
func (uc *UC) GetActiveDeployment(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdeploymentdto.GetActiveDeploymentReq,
) (*appdeploymentdto.GetActiveDeploymentResp, error) {
	deployments, _, err := uc.deploymentRepo.List(ctx, uc.db, req.AppID, nil,
		bunex.SelectColumns("id", "status", "created_at"),
		bunex.SelectWhereIn("deployment.status IN (?)",
			base.DeploymentStatusNotStarted, base.DeploymentStatusInProgress),
		bunex.SelectOrder("deployment.created_at ASC"),
		bunex.SelectLimit(activeDeploymentsMax),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(deployments) == 0 {
		return &appdeploymentdto.GetActiveDeploymentResp{}, nil
	}

	ids := make([]string, 0, len(deployments))
	for _, deployment := range deployments {
		ids = append(ids, deployment.ID)
	}
	if allowedAll, allowed := auth.AllowedDeployments(ids); !allowedAll {
		deployments = keepDeployments(deployments, allowed)
		if len(deployments) == 0 {
			return &appdeploymentdto.GetActiveDeploymentResp{}, nil
		}
	}

	// A deployment runs in a transaction of its own: until it ends, the row
	// still says not-started, and that it runs is in the cache.
	infos, err := uc.deploymentInfoRepo.MGet(ctx, ids)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appdeploymentdto.GetActiveDeploymentResp{Data: pickActiveDeployment(deployments, infos)}, nil
}

// pickActiveDeployment is the deployment running, if one is; else the next to
// run - the oldest queued. The deployments are the app's not ended, oldest
// first.
func pickActiveDeployment(
	deployments []*entity.Deployment,
	infos map[string]*cacheentity.DeploymentInfo,
) *appdeploymentdto.ActiveDeploymentResp {
	if len(deployments) == 0 {
		return nil
	}
	for _, deployment := range deployments {
		info := infos[deployment.ID]
		if deployment.Status == base.DeploymentStatusInProgress ||
			(info != nil && info.Status == base.DeploymentStatusInProgress) {
			return &appdeploymentdto.ActiveDeploymentResp{ID: deployment.ID, Status: base.DeploymentStatusInProgress}
		}
	}
	return &appdeploymentdto.ActiveDeploymentResp{ID: deployments[0].ID, Status: base.DeploymentStatusNotStarted}
}

// keepDeployments is the deployments whose ids are among those given.
func keepDeployments(deployments []*entity.Deployment, ids []string) []*entity.Deployment {
	keep := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		keep[id] = struct{}{}
	}
	kept := make([]*entity.Deployment, 0, len(deployments))
	for _, deployment := range deployments {
		if _, ok := keep[deployment.ID]; ok {
			kept = append(kept, deployment)
		}
	}
	return kept
}
