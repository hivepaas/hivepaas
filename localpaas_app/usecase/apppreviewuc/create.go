package apppreviewuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/service/apppreviewservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/apppreviewuc/apppreviewdto"
)

func (uc *UC) CreatePreview(
	ctx context.Context,
	auth *basedto.Auth,
	req *apppreviewdto.CreatePreviewReq,
) (_ *apppreviewdto.CreatePreviewResp, err error) {
	var createResp *apppreviewservice.CreatePreviewResp
	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		createResp, err = uc.appPreviewService.CreatePreview(ctx, db, &apppreviewservice.CreatePreviewReq{
			ProjectID:       req.ProjectID,
			AppID:           req.AppID,
			RepoRef:         req.RepoRef,
			CustomSubdomain: req.CustomSubdomain,
			NoStart:         req.NoStart,
			OnInitDeployment: func(deployment *entity.Deployment) error {
				// Set trigger for the deployment
				deployment.Trigger = &entity.AppDeploymentTrigger{
					Source:   base.DeploymentTriggerSourceUser,
					SourceID: auth.User.ID,
				}
				return nil
			},
		})
		if err != nil {
			return apperrors.New(err)
		}
		return nil
	})
	// Run the cleanup function
	if createResp != nil && createResp.OnCleanup != nil {
		_ = createResp.OnCleanup(err)
	}
	if err != nil {
		return nil, apperrors.New(err)
	}

	if createResp.DeploymentTask != nil {
		err = uc.taskQueue.ScheduleTask(ctx, createResp.DeploymentTask)
		if err != nil {
			return nil, apperrors.New(err)
		}
	}

	return &apppreviewdto.CreatePreviewResp{
		Data: &basedto.ObjectIDResp{ID: createResp.PreviewApp.ID},
	}, nil
}
