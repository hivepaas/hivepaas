package apppreviewserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice"
)

type createPreviewData struct {
	*apppreviewservice.CreatePreviewReq

	CalcRepoRef   string // normalized repo ref
	PullNumber    uint64
	CalcSubdomain string
	CalcAppName   string
	RandSuffix    string

	PreviewApp         *entity.App
	Deployment         *entity.Deployment
	DeploymentTask     *entity.Task
	DeploymentSettings *entity.AppDeploymentSettings

	CloneDBAppsData           map[string]*cloneDBAppData
	CloneDBAppsEnvRefReplacer *strings.Replacer
}

func (s *service) CreatePreview(
	ctx context.Context,
	db database.Tx,
	req *apppreviewservice.CreatePreviewReq,
) (_ *apppreviewservice.CreatePreviewResp, err error) {
	data := &createPreviewData{
		CreatePreviewReq: req,
	}

	err = s.loadAppDataForCreatingPreview(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	var cloneResp *appcloneservice.AppCloneResp
	defer func() {
		if rev := recover(); rev != nil {
			err = errors.Join(err, apperrors.ErrPanic)
		}
		if cloneResp != nil && cloneResp.OnCleanup != nil { // Run the cleanup function
			_ = cloneResp.OnCleanup(err)
			cloneResp.OnCleanup = nil
		}
	}()

	// When related DB apps are configured to clone for the previews
	err = s.cloneDBApps(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	cloneResp, err = s.appCloneService.CloneApp(ctx, db, &appcloneservice.AppCloneReq{
		SrcApp: data.App,
		OnCloneApp: func(targetApp, srcApp *entity.App) error {
			data.PreviewApp = targetApp
			return s.onCloneApp(targetApp, srcApp, data)
		},
		OnCloneSetting: func(targetApp, srcApp *entity.App, setting *entity.Setting) (*entity.Setting, error) {
			return s.onCloneAppSetting(ctx, db, setting, data)
		},
		OnCloneService: func(targetApp, srcApp *entity.App, targetSvc, srcSvc *swarm.Service) error {
			return s.onCloneAppService(targetSvc, srcSvc, data)
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.createDeploymentAndTask(ctx, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.persistAppPreviewData(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &apppreviewservice.CreatePreviewResp{
		PreviewApp:     data.PreviewApp,
		Deployment:     data.Deployment,
		DeploymentTask: data.DeploymentTask,
		OnCleanup:      cloneResp.OnCleanup,
	}, apperrors.Wrap(err)
}

func (s *service) loadAppDataForCreatingPreview(
	ctx context.Context,
	db database.IDB,
	data *createPreviewData,
) (err error) {
	app, err := s.appService.LoadApp(ctx, db, data.App.ProjectID, data.App.ID, true, true,
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	// The app must not be a child app
	if app.IsChildApp() {
		return apperrors.Wrap(apperrors.ErrActionNotAllowed).WithMsgLog("child app cannot have a preview")
	}

	deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
	if deploymentSetting == nil {
		return apperrors.NewNotFound("Deployment settings")
	}
	deploymentSettings := deploymentSetting.MustAsAppDeploymentSettings()
	if deploymentSettings.ActiveMethod != base.DeploymentMethodRepo || deploymentSettings.RepoSource == nil {
		return apperrors.Wrap(apperrors.ErrDeploymentMethodRepoRequired)
	}
	data.App = app

	data.RandSuffix = gofn.RandTokenAsHex(4) //nolint:mnd
	data.CalcRepoRef, data.PullNumber, err = githelper.NormalizePullRef(data.RepoRef)
	if err != nil {
		data.CalcRepoRef = string(githelper.NormalizeRepoRef(data.RepoRef))
		data.PullNumber = 0
	}
	data.CalcSubdomain = data.CustomSubdomain
	if data.CalcSubdomain == "" && data.PullNumber > 0 {
		data.CalcSubdomain = fmt.Sprintf("pr-%v", data.PullNumber)
	}
	if data.CalcSubdomain == "" {
		data.CalcSubdomain = data.RandSuffix
	}
	if data.PullNumber > 0 {
		data.CalcAppName = fmt.Sprintf("pr-%v", data.PullNumber)
	}
	if data.CalcAppName == "" {
		data.CalcAppName = data.CalcSubdomain
	}

	previewApp, err := s.GetPreview(ctx, db, app.ID, data.CalcRepoRef, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if previewApp != nil {
		return apperrors.NewAlreadyExist("Preview app")
	}

	return nil
}

func (s *service) createDeploymentAndTask(
	_ context.Context,
	data *createPreviewData,
) (err error) {
	previewApp := data.PreviewApp
	deployment, deploymentTask, err := s.appDeploymentService.CreateDeploymentAndTask(
		previewApp, data.DeploymentSettings)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if data.OnInitDeployment != nil {
		if err = data.OnInitDeployment(deployment); err != nil {
			return apperrors.Wrap(err)
		}
	}
	if data.OnDeploymentTask != nil {
		if err = data.OnDeploymentTask(deploymentTask); err != nil {
			return apperrors.Wrap(err)
		}
	}

	data.Deployment = deployment
	data.DeploymentTask = deploymentTask
	return nil
}

func (s *service) persistAppPreviewData(
	ctx context.Context,
	db database.IDB,
	data *createPreviewData,
) (err error) {
	err = s.deploymentRepo.Upsert(ctx, db, data.Deployment,
		entity.DeploymentUpsertingConflictCols, entity.DeploymentUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = s.taskRepo.Upsert(ctx, db, data.DeploymentTask,
		entity.TaskUpsertingConflictCols, entity.TaskUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// If there are cloned db apps, we will add some res-links for them as logical-child-apps
	resLinks := make([]*entity.ResLink, 0, len(data.CloneDBAppsData))
	timeNow := time.Now()
	for _, appData := range data.CloneDBAppsData {
		resLinks = append(resLinks, &entity.ResLink{
			SrcType:   base.ResourceTypeApp,
			SrcID:     data.PreviewApp.ID,
			DstType:   base.ResourceTypeLogicalChildApp,
			DstID:     appData.CloneResp.TargetApp.ID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}
	err = s.resLinkRepo.UpsertMulti(ctx, db, resLinks,
		entity.ResLinkUpsertingConflictCols, entity.ResLinkUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
