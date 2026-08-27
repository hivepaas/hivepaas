package apppreviewuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
		data := &createPreviewData{}
		err := uc.loadAppForCreatePreview(ctx, db, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		cloneDBApps := data.PreviewSettings.AutoCloneApps
		if req.CloneDBApps != nil {
			cloneDBApps = *req.CloneDBApps
		}

		previewTask, err = uc.appPreviewService.CreateAppPreviewTask(data.App, &entity.TaskAppPreviewArgs{
			ParentApp:       entity.ObjectID{ID: data.App.ID},
			RepoRef:         req.RepoRef,
			CustomSubdomain: req.CustomSubdomain,
			NoStart:         req.NoStart,
			CloneDBApps:     cloneDBApps,
			Trigger: &entity.AppDeploymentTrigger{
				Source:   base.DeploymentTriggerSourceUser,
				SourceID: auth.User.ID,
			},
		})
		if err != nil {
			return hperrors.Wrap(err)
		}

		err = uc.taskRepo.Upsert(ctx, db, previewTask,
			entity.TaskUpsertingConflictCols, entity.TaskUpsertingUpdateCols)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if previewTask != nil {
		err = uc.taskQueue.ScheduleTask(ctx, previewTask)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	return &apppreviewdto.CreatePreviewResp{
		Data: &basedto.ObjectIDResp{ID: previewTask.ID},
	}, nil
}

type createPreviewData struct {
	App             *entity.App
	PreviewSettings *entity.AppFeaturePreviewSettings
}

func (uc *UC) loadAppForCreatePreview(
	ctx context.Context,
	db database.IDB,
	req *apppreviewdto.CreatePreviewReq,
	data *createPreviewData,
) error {
	app, featureSettings, err := uc.appService.LoadAppWithFeatureSettings(ctx, db, req.ProjectID, req.AppID,
		true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	// The app must not be a child app
	if app.IsChildApp() {
		return hperrors.Wrap(hperrors.ErrActionNotAllowed).WithMsgLog("child app cannot have a preview")
	}
	data.App = app

	previewSettings := featureSettings.PreviewSettings
	if previewSettings == nil || !previewSettings.Enabled {
		return hperrors.Wrap(hperrors.ErrFeatureDisabled).WithParam("Name", "app preview")
	}

	// Check if preview already exists for this repo ref
	calcRepoRef, _, err := githelper.NormalizePullRef(req.RepoRef)
	if err != nil {
		calcRepoRef = string(githelper.NormalizeRepoRef(req.RepoRef))
	}
	existingPreview, err := uc.appPreviewService.GetPreview(ctx, db, app.ID, calcRepoRef, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if existingPreview != nil {
		return hperrors.NewAlreadyExist("Preview app")
	}
	data.PreviewSettings = previewSettings

	return nil
}
