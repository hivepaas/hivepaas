package appuc

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/apphelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

const (
	dockerImageInit    = "busybox:latest"
	dockerImageInitDev = "crccheck/hello-world:latest"
)

func (uc *UC) CreateApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.CreateAppReq,
) (resp *appdto.CreateAppResp, err error) {
	resp = &appdto.CreateAppResp{}
	var createdApp *entity.App

	defer func() {
		if rec := recover(); rec != nil {
			err = errors.Join(err, apperrors.ErrPanic)
		}
		if err != nil && createdApp != nil && createdApp.ServiceID != "" {
			_ = uc.clusterService.ServiceRemove(ctx, createdApp.ServiceID, clusterservice.ItemRemovalRetryMax, 0)
		}
	}()

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		appData := &createAppData{}
		err := uc.loadAppData(ctx, db, req, appData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		err = uc.preparePersistingApp(ctx, db, req, appData, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		createdApp = persistingData.UpsertingApps[0]
		resp.Data = &basedto.ObjectIDResp{ID: createdApp.ID}

		// Create a service in docker for the app
		res, err := uc.dockerManager.ServiceCreate(ctx, appData.ServiceSpec)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if res.ID == "" { // should never happen
			return apperrors.Wrap(apperrors.ErrInfraInternal).
				WithNTParam("Error", "empty service ID returned")
		}
		createdApp.ServiceID = res.ID

		return uc.persistData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, nil
}

type createAppData struct {
	Project      *entity.Project
	ProjectEnv   *entity.ProjectEnv
	AppGlobalKey string
	AppKey       string
	ServiceSpec  *swarm.ServiceSpec
}

func (uc *UC) loadAppData(
	ctx context.Context,
	db database.IDB,
	req *appdto.CreateAppReq,
	data *createAppData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ProjectID,
		bunex.SelectFor("UPDATE OF project"),
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		bunex.SelectRelation("ProjectEnvs",
			bunex.SelectWhere("project_env.id = ?", req.ProjectEnvID),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if project.Status != base.ProjectStatusActive {
		return apperrors.Wrap(apperrors.ErrProjectInactive).WithNTParam("Name", project.Name)
	}
	if len(project.ProjectEnvs) == 0 {
		return apperrors.Wrap(apperrors.ErrProjectEnvNotFound).WithParam("Name", req.ProjectEnvID)
	}
	projectEnv := project.ProjectEnvs[0]
	if projectEnv.Status != base.ProjectStatusActive {
		return apperrors.Wrap(apperrors.ErrProjectEnvInactive).WithNTParam("Project", project.Name).
			WithNTParam("Env", projectEnv.Name)
	}

	data.Project = project
	data.ProjectEnv = projectEnv
	data.AppKey = projecthelper.CalcAppKey(req.Name)
	data.AppGlobalKey = projecthelper.CalcAppGlobalKey(project.Key, data.AppKey, projectEnv.Key)

	// App keys must be unique globally
	conflictApp, err := uc.appRepo.GetByGlobalKey(ctx, db, "", data.AppGlobalKey, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if conflictApp != nil {
		return apperrors.NewAlreadyExist("App").
			WithMsgLog("app unique key '%s' already exists", data.AppGlobalKey)
	}

	// Create local network for the app to attach
	_, _, err = uc.networkService.GetOrCreateProjectNetwork(ctx, db, project, projectEnv.Key)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

type persistingAppData struct {
	appservice.PersistingAppData
}

func (uc *UC) preparePersistingApp(
	ctx context.Context,
	db database.IDB,
	req *appdto.CreateAppReq,
	data *createAppData,
	persistingData *persistingAppData,
) error {
	timeNow := timeutil.NowUTC()
	project := data.Project
	app := &entity.App{
		ID:           gofn.Must(ulid.NewStringULID()),
		ProjectID:    project.ID,
		ProjectEnvID: data.ProjectEnv.ID,
		Key:          data.AppKey,
		GlobalKey:    data.AppGlobalKey,
		CreatedAt:    timeNow,
	}

	uc.preparePersistingAppBase(app, req.AppBaseReq, timeNow, persistingData)
	uc.preparePersistingAppTags(app, req.Tags, 0, persistingData)
	uc.preparePersistingAppSettingsDefault(app, timeNow, persistingData)

	err := uc.preparePersistingAppService(ctx, db, app, data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (uc *UC) preparePersistingAppBase(
	app *entity.App,
	req *appdto.AppBaseReq,
	timeNow time.Time,
	persistingData *persistingAppData,
) {
	app.Name = req.Name
	app.Status = req.Status
	app.Note = req.Note
	app.UpdatedAt = timeNow

	persistingData.UpsertingApps = append(persistingData.UpsertingApps, app)
}

func (uc *UC) preparePersistingAppTags(
	app *entity.App,
	tags []string,
	startIndex int,
	persistingData *persistingAppData,
) {
	index := startIndex
	for _, tag := range tags {
		persistingData.UpsertingTags = append(persistingData.UpsertingTags,
			&entity.Tag{
				ObjectID: app.ID,
				Tag:      tag,
				Index:    index,
			})
		index++
	}
}

func (uc *UC) preparePersistingAppService(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	data *createAppData,
) error {
	isDevEnv := config.Current.IsDevEnv()

	appInfo := &apphelper.AppInfo{
		Name: app.Name,
		Key:  app.Key,
		Env:  data.ProjectEnv.Name,
	}

	service := &swarm.Service{
		Spec: swarm.ServiceSpec{
			Mode: swarm.ServiceMode{
				Replicated: &swarm.ReplicatedService{
					Replicas: new(uint64(1)),
				},
			},
			Annotations: swarm.Annotations{
				Name: app.GlobalKey,
				Labels: map[string]string{
					appservice.LabelAppNamespace: data.Project.Key,
					appservice.LabelAppInfo:      apphelper.CalcAppInfoLabel(appInfo),
				},
			},
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: &swarm.ContainerSpec{
					Image:    gofn.If(isDevEnv, dockerImageInitDev, dockerImageInit),
					Command:  gofn.If(isDevEnv, nil, []string{"sleep", "infinity"}),
					Hostname: app.Key,
					Init:     new(true), // default to use `tini`
				},
				Networks: []swarm.NetworkAttachmentConfig{
					{
						Target:  uc.networkService.GetProjectNetworkName(data.Project, data.ProjectEnv.Name),
						Aliases: []string{app.Key},
					},
				},
				LogDriver: &swarm.Driver{
					// Default driver is `json-file`, but Docker recommends `local`
					// See: https://docs.docker.com/engine/logging/configure/
					Name: "local",
					Options: map[string]string{
						"max-size": "50m",
						"max-file": "20",
						"compress": "true",
					},
				},
			},
		},
	}

	_, err := uc.placementService.ApplyPlacementSettings(ctx, db, &placementservice.ApplyPlacementSettingsReq{
		App:                app,
		Service:            service,
		SkipSavingToDocker: true,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	data.ServiceSpec = &service.Spec
	return nil
}

func (uc *UC) preparePersistingAppSettingsDefault(
	app *entity.App,
	timeNow time.Time,
	persistingData *persistingAppData,
) {
	// Init empty http settings
	httpSettings := &entity.AppHttpSettings{}
	dbHttpSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeApp,
		Type:      base.SettingTypeAppHttp,
		Status:    base.SettingStatusActive,
		ObjectID:  app.ID,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	dbHttpSetting.MustSetData(httpSettings)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, dbHttpSetting)

	// Init feature settings
	featureSettings := &entity.AppFeatureSettings{}
	entity.InitAppFeatureSettingsDefault(featureSettings)
	dbFeatureSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeApp,
		Type:      base.SettingTypeAppFeatures,
		Status:    base.SettingStatusActive,
		ObjectID:  app.ID,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	dbFeatureSetting.MustSetData(featureSettings)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, dbFeatureSetting)
}

func (uc *UC) persistData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingAppData,
) error {
	err := uc.appService.PersistAppData(ctx, db, &persistingData.PersistingAppData)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
