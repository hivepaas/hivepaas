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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type createPreviewData struct {
	*apppreviewservice.CreatePreviewReq

	Args   *entity.TaskAppPreviewArgs
	Output *entity.TaskAppPreviewOutput

	CalcRepoRef   string // normalized repo ref
	PullNumber    uint64
	CalcSubdomain string
	CalcAppName   string
	RandSuffix    string

	FeatureSettings    *entity.AppFeatureSettings
	PreviewApp         *entity.App
	Deployment         *entity.Deployment
	DeploymentTask     *entity.Task
	DeploymentSettings *entity.AppDeploymentSettings

	CloneDBApps               []*entity.App
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
		Output:           &entity.TaskAppPreviewOutput{},
	}
	s.initLogStore(data)
	defer func() {
		ctx = context.WithoutCancel(ctx)
		_ = s.saveLogs(ctx, db, data)
	}()

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

	// Preview app will be cloned from the current app
	cloneTask := &entity.Task{
		ID:       "fake-task-id",
		Scope:    base.ObjectScopeApp,
		ObjectID: data.App.ID,
		Type:     base.TaskTypeAppClone,
	}
	cloneTask.MustSetArgs(&entity.TaskAppCloneArgs{SrcApp: entity.ObjectID{ID: data.App.ID}})

	cloneResp, err = s.appCloneService.CloneApp(ctx, db, &appcloneservice.AppCloneReq{
		TaskExecData: &queue.TaskExecData{
			Task:       cloneTask,
			RefObjects: data.RefObjects,
			LogStore:   data.LogStore,
		},
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

	// Run custom commands
	err = s.runCommands(ctx, db, data)
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

	if s.taskQueue != nil && data.DeploymentTask != nil {
		data.OnPostTransaction(func() { //nolint:contextcheck
			_ = s.taskQueue.ScheduleTask(context.Background(), data.DeploymentTask)
		})
	}

	// Save result to the task's output
	data.Output.PreviewApp = entity.ObjectID{ID: data.PreviewApp.ID}
	data.Output.Deployment = entity.ObjectID{ID: data.Deployment.ID}
	_ = data.Task.SetOutput(data.Output)

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
	taskArgs, err := data.Task.ArgsAsAppPreview()
	if err != nil {
		return apperrors.Wrap(err)
	}
	if taskArgs == nil || taskArgs.ParentApp.ID == "" {
		return apperrors.NewNotFound("Parent app ID in task args")
	}
	data.Args = taskArgs
	if taskArgs.Trigger != nil {
		data.OnInitDeployment = func(deployment *entity.Deployment) error {
			deployment.Trigger = taskArgs.Trigger
			return nil
		}
	}

	app, err := s.appService.LoadApp(ctx, db, "", taskArgs.ParentApp.ID, true, true,
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeAppDeployment,
				base.SettingTypeAppFeatures),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	// The app must not be a child app
	if app.IsChildApp() {
		return apperrors.Wrap(apperrors.ErrActionNotAllowed).WithMsgLog("child app cannot have a preview")
	}
	data.App = app

	if featureSetting := app.GetSettingByType(base.SettingTypeAppFeatures); featureSetting != nil {
		data.FeatureSettings = featureSetting.MustAsAppFeatureSettings()
	}

	if data.FeatureSettings != nil && data.FeatureSettings.PreviewSettings != nil {
		previewSettings := data.FeatureSettings.PreviewSettings
		if !previewSettings.Enabled {
			return apperrors.Wrap(apperrors.ErrFeatureDisabled).WithParam("Name", "app preview")
		}
		var cloningAppIDs []string
		if taskArgs.CloneDBApps || (previewSettings.AutoCloneApps && !taskArgs.SkipCloningDBApps) {
			cloningAppIDs = previewSettings.AppsToClone.ToIDStringSlice()
		}
		cloningApps, err := s.appService.LoadAppsSkipMissing(ctx, db, app.Project.ID, cloningAppIDs, true, false,
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
	}

	deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
	if deploymentSetting == nil {
		return apperrors.NewNotFound("Deployment settings")
	}
	deploymentSettings := deploymentSetting.MustAsAppDeploymentSettings()
	if deploymentSettings.ActiveMethod != base.DeploymentMethodRepo || deploymentSettings.RepoSource == nil {
		return apperrors.Wrap(apperrors.ErrDeploymentMethodRepoRequired)
	}

	data.RandSuffix = gofn.RandTokenAsHex(4) //nolint:mnd
	data.CalcRepoRef, data.PullNumber, err = githelper.NormalizePullRef(taskArgs.RepoRef)
	if err != nil {
		data.CalcRepoRef = string(githelper.NormalizeRepoRef(taskArgs.RepoRef))
		data.PullNumber = 0
	}
	data.CalcSubdomain = taskArgs.CustomSubdomain
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

func (s *service) initLogStore(data *createPreviewData) {
	if data.LogStore != nil {
		return
	}
	if data.Task != nil && s.redisClient != nil {
		data.LogStore = tasklog.NewRemoteStore(fmt.Sprintf("task:%s:log", data.Task.ID), s.redisClient)
		data.LogStore.SetOnFlush(tasklog.DefaultMaxSize, func(ctx context.Context, frames []*tasklog.LogFrame) error {
			return s.saveLogFramesToDB(ctx, s.db, data.Task.ID, frames)
		})
	} else {
		data.LogStore = tasklog.NewNullStore()
	}
}

func (s *service) saveLogs(
	ctx context.Context,
	db database.IDB,
	data *createPreviewData,
) error {
	logStore := data.LogStore
	if logStore == nil || data.Task == nil {
		return nil
	}

	logFrames, err := logStore.GetData(ctx, 0)
	if err != nil {
		return apperrors.Wrap(err)
	}
	_ = logStore.Close() //nolint

	return s.saveLogFramesToDB(ctx, db, data.Task.ID, logFrames)
}

func (s *service) saveLogFramesToDB(
	ctx context.Context,
	db database.IDB,
	taskID string,
	logFrames []*tasklog.LogFrame,
) error {
	if s.taskLogRepo == nil || len(logFrames) == 0 {
		return nil
	}
	for _, chunk := range gofn.Chunk(logFrames, 10000) { //nolint
		taskLogs := make([]*entity.TaskLog, 0, len(chunk))
		for _, logFrame := range chunk {
			taskLogs = append(taskLogs, &entity.TaskLog{
				TaskID: taskID,
				Type:   logFrame.Type,
				Data:   logFrame.Data,
				Ts:     logFrame.Ts,
			})
		}
		err := s.taskLogRepo.InsertMulti(ctx, db, taskLogs)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}
