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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/slugify"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
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
) (*appdto.CreateAppResp, error) {
	var appData *createAppData
	var persistingData *persistingAppData
	var createdApp *entity.App
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		appData = &createAppData{}
		err := uc.loadAppData(ctx, db, req, appData)
		if err != nil {
			return apperrors.New(err)
		}

		persistingData = &persistingAppData{}
		uc.preparePersistingApp(req, appData, persistingData)

		createdApp = persistingData.UpsertingApps[0]

		// Create a service in docker for the app
		res, err := uc.dockerManager.ServiceCreate(ctx, appData.ServiceSpec)
		if err != nil {
			return apperrors.New(err)
		}
		if res.ID == "" { // should never happen
			return apperrors.New(apperrors.ErrInfraInternal).
				WithNTParam("Error", "empty service ID returned")
		}
		createdApp.ServiceID = res.ID

		return uc.persistData(ctx, db, persistingData)
	})
	if err != nil {
		// Transaction fails, but service is created in docker, need to delete it
		if createdApp != nil && createdApp.ServiceID != "" {
			_ = uc.clusterService.ServiceRemove(ctx, createdApp.ServiceID, clusterservice.ItemRemovalRetryMax, 0)
		}
		return nil, apperrors.New(err)
	}

	return &appdto.CreateAppResp{
		Data: &basedto.ObjectIDResp{ID: createdApp.ID},
	}, nil
}

type createAppData struct {
	Project         *entity.Project
	AppKey          string
	AppLocalKey     string
	LocalNetAliases []string
	ServiceSpec     *swarm.ServiceSpec
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
	)
	if err != nil {
		return apperrors.New(err)
	}
	if project.Status != base.ProjectStatusActive {
		return apperrors.New(apperrors.ErrProjectInactive).WithNTParam("Name", project.Name)
	}
	data.Project = project

	appSlug := slugify.SlugifyAsKey(req.Name)
	data.AppLocalKey = appSlug
	if req.Env != "" {
		data.AppLocalKey += "_" + projecthelper.CalcProjectEnvKey(req.Env)
	}
	data.AppKey = project.Key + "_" + data.AppLocalKey
	data.LocalNetAliases = append(data.LocalNetAliases, appSlug)

	// App keys must be unique globally
	conflictApp, err := uc.appRepo.GetByKey(ctx, db, "", data.AppKey, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.New(err)
	}
	if conflictApp != nil {
		return apperrors.NewAlreadyExist("App").
			WithMsgLog("app key '%s' already exists", data.AppKey)
	}

	// Create local network for the app to attach
	_, _, err = uc.networkService.GetOrCreateProjectNetwork(ctx, db, project, req.Env)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}

type persistingAppData struct {
	appservice.PersistingAppData
}

func (uc *UC) preparePersistingApp(
	req *appdto.CreateAppReq,
	data *createAppData,
	persistingData *persistingAppData,
) {
	timeNow := timeutil.NowUTC()
	project := data.Project
	app := &entity.App{
		ID:        gofn.Must(ulid.NewStringULID()),
		ProjectID: project.ID,
		Key:       data.AppKey,
		LocalKey:  data.AppLocalKey,
		CreatedAt: timeNow,
	}
	uc.preparePersistingAppBase(app, req.AppBaseReq, timeNow, persistingData)
	uc.preparePersistingAppTags(app, req.Tags, 0, persistingData)
	uc.preparePersistingAppSettingsDefault(app, timeNow, data, persistingData)
}

func (uc *UC) preparePersistingAppBase(
	app *entity.App,
	req *appdto.AppBaseReq,
	timeNow time.Time,
	persistingData *persistingAppData,
) {
	app.Name = req.Name
	app.Status = req.Status
	app.Env = req.Env
	app.Note = req.Note
	app.UpdatedAt = timeNow

	persistingData.UpsertingApps = append(persistingData.UpsertingApps, app)
}

func (uc *UC) preparePersistingAppTags(
	app *entity.App,
	tags []string,
	startDisplayOrder int,
	persistingData *persistingAppData,
) {
	displayOrder := startDisplayOrder
	for _, tag := range tags {
		persistingData.UpsertingTags = append(persistingData.UpsertingTags,
			&entity.AppTag{
				AppID:        app.ID,
				Tag:          tag,
				DisplayOrder: displayOrder,
			})
		displayOrder++
	}
}

func (uc *UC) preparePersistingAppSettingsDefault(
	app *entity.App,
	timeNow time.Time,
	data *createAppData,
	persistingData *persistingAppData,
) {
	isDevEnv := config.Current.IsDevEnv()
	serviceSpec := &swarm.ServiceSpec{
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: new(uint64(1)),
			},
		},
		Annotations: swarm.Annotations{
			Name: app.Key,
			Labels: map[string]string{
				appservice.LabelAppNamespace: data.Project.Key,
				appservice.LabelAppName:      app.Name,
				appservice.LabelAppEnv:       app.Env,
			},
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:    gofn.If(isDevEnv, dockerImageInitDev, dockerImageInit),
				Command:  gofn.If(isDevEnv, nil, []string{"sleep", "infinity"}),
				Hostname: app.LocalKey,
				Init:     new(true), // default to use `tini`
			},
			Networks: []swarm.NetworkAttachmentConfig{
				{
					Target:  uc.networkService.GetProjectNetworkName(data.Project, app.Env),
					Aliases: data.LocalNetAliases,
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
	}
	data.ServiceSpec = serviceSpec

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
		return apperrors.New(err)
	}
	return nil
}
