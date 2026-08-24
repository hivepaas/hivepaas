package apppreviewuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apppreviewuc/apppreviewdto"
)

func (uc *UC) CreatePreview(
	ctx context.Context,
	auth *basedto.Auth,
	req *apppreviewdto.CreatePreviewReq,
) (*apppreviewdto.CreatePreviewResp, error) {
	var previewTask *entity.Task
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		app, err := uc.loadAppForCreatePreview(ctx, db, req)
		if err != nil {
			return apperrors.Wrap(err)
		}

		previewTask, err = uc.appPreviewService.CreateAppPreviewTask(app, &entity.TaskAppPreviewArgs{
			ParentApp:         entity.ObjectID{ID: app.ID},
			RepoRef:           req.RepoRef,
			CustomSubdomain:   req.CustomSubdomain,
			NoStart:           req.NoStart,
			CloneDBApps:       req.CloneDBApps,
			SkipCloningDBApps: req.SkipCloningDBApps,
			Trigger: &entity.AppDeploymentTrigger{
				Source:   base.DeploymentTriggerSourceUser,
				SourceID: auth.User.ID,
			},
		})
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.taskRepo.Upsert(ctx, db, previewTask,
			entity.TaskUpsertingConflictCols, entity.TaskUpsertingUpdateCols)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if previewTask != nil {
		err = uc.taskQueue.ScheduleTask(ctx, previewTask)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	return &apppreviewdto.CreatePreviewResp{
		Data: &basedto.ObjectIDResp{ID: previewTask.ID},
	}, nil
}

func (uc *UC) loadAppForCreatePreview(
	ctx context.Context,
	db database.IDB,
	req *apppreviewdto.CreatePreviewReq,
) (*entity.App, error) {
	app, featureSettings, err := uc.appService.LoadAppWithFeatureSettings(ctx, db, req.ProjectID, req.AppID,
		true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	// The app must not be a child app
	if app.IsChildApp() {
		return nil, apperrors.Wrap(apperrors.ErrActionNotAllowed).WithMsgLog("child app cannot have a preview")
	}

	previewSettings := featureSettings.PreviewSettings
	if previewSettings == nil || !previewSettings.Enabled {
		return nil, apperrors.Wrap(apperrors.ErrFeatureDisabled).WithParam("Name", "app preview")
	}

	// Check if preview already exists for this repo ref
	calcRepoRef, _, err := githelper.NormalizePullRef(req.RepoRef)
	if err != nil {
		calcRepoRef = string(githelper.NormalizeRepoRef(req.RepoRef))
	}
	existingPreview, err := uc.appPreviewService.GetPreview(ctx, db, app.ID, calcRepoRef, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.Wrap(err)
	}
	if existingPreview != nil {
		return nil, apperrors.NewAlreadyExist("Preview app")
	}

	return app, nil
}
