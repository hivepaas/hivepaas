package appuc

import (
	"context"
	"errors"
	"time"

	"github.com/docker/docker/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/slugify"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	appKeyMaxLen = 100
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
			return apperrors.Wrap(err)
		}

		persistingData = &persistingAppData{}
		uc.preparePersistingApp(req, appData, persistingData)

		createdApp = persistingData.UpsertingApps[0]

		// Create a service in docker for the app
		res, err := uc.dockerManager.ServiceCreate(ctx, appData.ServiceSpec)
		if err != nil {
			return apperrors.Wrap(err)
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
			_ = uc.dockerManager.ServiceRemove(ctx, createdApp.ServiceID)
		}
		return nil, apperrors.Wrap(err)
	}

	return &appdto.CreateAppResp{
		Data: &basedto.ObjectIDResp{ID: createdApp.ID},
	}, nil
}

type createAppData struct {
	Project     *entity.Project
	AppKey      string
	ServiceSpec *swarm.ServiceSpec
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
		return apperrors.Wrap(err)
	}
	if project.Status != base.ProjectStatusActive {
		return apperrors.New(apperrors.ErrProjectInactive).WithNTParam("Name", project.Name)
	}
	data.Project = project

	data.AppKey = project.Key + "_" + slugify.SlugifyEx(req.Name, nil, appKeyMaxLen)

	// App keys must be unique globally
	conflictApp, err := uc.appRepo.GetByKey(ctx, db, "", data.AppKey, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if conflictApp != nil {
		return apperrors.NewAlreadyExist("App").
			WithMsgLog("app key '%s' already exists", data.AppKey)
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
		CreatedAt: timeNow,
	}
	app.ResetToken()

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
	serviceSpec := &swarm.ServiceSpec{
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: new(uint64(1)),
			},
		},
		Annotations: swarm.Annotations{
			Name: app.Key,
			Labels: map[string]string{
				docker.StackLabelNamespace: data.Project.Key,
			},
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:    "crccheck/hello-world:latest", // TODO: we can use busybox:latest
				Hostname: app.Key,
			},
			Networks: []swarm.NetworkAttachmentConfig{
				{
					Target: data.Project.GetDefaultNetworkName(),
				},
			},
		},
	}
	data.ServiceSpec = serviceSpec

	// Init empty http settings
	httpSettings := &entity.AppHttpSettings{}
	dbHttpSettings := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.SettingScopeApp,
		Type:      base.SettingTypeAppHttp,
		Status:    base.SettingStatusActive,
		ObjectID:  app.ID,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	dbHttpSettings.MustSetData(httpSettings)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, dbHttpSettings)
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
