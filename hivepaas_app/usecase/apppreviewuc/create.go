package apppreviewuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apppreviewuc/apppreviewdto"
)

func (uc *UC) CreatePreview(
	ctx context.Context,
	auth *basedto.Auth,
	req *apppreviewdto.CreatePreviewReq,
) (_ *apppreviewdto.CreatePreviewResp, err error) {
	var createResp *apppreviewservice.CreatePreviewResp
	defer func() {
		if (err != nil || recover() != nil) && createResp != nil && createResp.OnCleanup != nil {
			_ = createResp.OnCleanup(gofn.Coalesce(err, apperrors.ErrPanic))
		}
	}()

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		// DO NOT use SELECT FOR UPDATE when load the parent app.
		// Creating a preview may take time, so we don't lock the parent app.
		// However, after creating, we check the app status again. If it's not active and valid,
		// the preview app will be deleted.

		data := &createAppPreviewData{}
		err = uc.loadAppForCreatePreview(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		createResp, err = uc.appPreviewService.CreatePreview(ctx, db, &apppreviewservice.CreatePreviewReq{
			App:             data.App,
			RepoRef:         req.RepoRef,
			CustomSubdomain: req.CustomSubdomain,
			NoStart:         req.NoStart,
			CloneDBApps:     data.CloneDBApps,
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
			return apperrors.Wrap(err)
		}

		// Ensure app valid when we complete creating the preview
		err = uc.appService.EnsureAppActive(ctx, db, data.App, false, false)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if createResp.DeploymentTask != nil {
		err = uc.taskQueue.ScheduleTask(ctx, createResp.DeploymentTask)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	return &apppreviewdto.CreatePreviewResp{
		Data: &basedto.ObjectIDResp{ID: createResp.PreviewApp.ID},
	}, nil
}

type createAppPreviewData struct {
	App         *entity.App
	CloneDBApps []*entity.App
}

func (uc *UC) loadAppForCreatePreview(
	ctx context.Context,
	db database.IDB,
	req *apppreviewdto.CreatePreviewReq,
	data *createAppPreviewData,
) (err error) {
	// NOTE: DO NOT use SELECT FOR UPDATE when load the parent app
	app, featureSettings, err := uc.appService.LoadAppWithFeatureSettings(ctx, db, req.ProjectID, req.AppID,
		true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	previewSettings := featureSettings.PreviewSettings
	if previewSettings == nil || !previewSettings.Enabled {
		return apperrors.Wrap(apperrors.ErrFeatureDisabled).WithParam("Name", "app preview")
	}
	data.App = app

	var cloningAppIDs []string
	if req.CloneDBApps || (previewSettings.AutoCloneApps && !req.SkipCloningDBApps) {
		cloningAppIDs = previewSettings.AppsToClone.ToIDStringSlice()
	}

	cloningApps, err := uc.appService.LoadAppsSkipMissing(ctx, db, app.Project.ID, cloningAppIDs, true, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.CloneDBApps = cloningApps

	return nil
}
