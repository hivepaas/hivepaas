package appcopyserviceimpl

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/slugify"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcopyservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
)

type appCopyData struct {
	*appcopyservice.AppCopyReq

	TargetApp     *entity.App
	SrcService    *swarm.Service
	TargetService *swarm.Service
	TargetSecrets []*entity.SwarmSecretRef
	TargetConfig  []*entity.SwarmConfigRef

	CopiedSettings []*entity.Setting
	RefObjects     *entity.RefObjects

	TimeNow time.Time
}

func (s *service) CopyApp(
	ctx context.Context,
	db database.Tx,
	req *appcopyservice.AppCopyReq,
) (resp *appcopyservice.AppCopyResp, err error) {
	resp = &appcopyservice.AppCopyResp{}
	data := &appCopyData{
		AppCopyReq: req,
		TimeNow:    timeutil.NowUTC(),
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.Join(err, apperrors.NewPanic(r))
		}
		_ = s.cleanupOnFail(ctx, data, err)
	}()

	err = s.copyApp(ctx, db, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	err = s.copyAppSettings(ctx, db, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	err = s.copySwarmService(ctx, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	err = s.createSwarmService(ctx, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	err = s.persistAppData(ctx, db, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	err = s.applyEnvVars(ctx, db, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	err = s.applySwarmConfigFiles(ctx, db, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	err = s.applySwarmSecrets(ctx, db, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	err = s.applyAppHttpSettings(ctx, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	err = s.applySchedJobSettings(ctx, db, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	err = s.applyFinalContainerSettings(ctx, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp.TargetApp = data.TargetApp
	resp.TargetService = data.TargetService
	resp.OnCleanup = func(e error) error {
		return s.cleanupOnFail(ctx, data, e)
	}
	return resp, nil
}

func (s *service) copyApp(
	ctx context.Context,
	db database.IDB,
	data *appCopyData,
) (err error) {
	timeNow := timeutil.NowUTC()
	targetApp := &entity.App{
		ID:        gofn.Must(ulid.NewStringULID()),
		ProjectID: data.TargetProject.ID,
		Project:   data.TargetProject,
		Status:    base.AppStatusActive,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	data.TargetApp = targetApp

	err = data.OnCopyApp(targetApp, data.SrcApp)
	if err != nil {
		return apperrors.New(err)
	}

	targetApp.LocalKey = slugify.SlugifyAsKey(targetApp.Name)
	if targetApp.ParentApp != nil {
		parentAppKey := targetApp.ParentApp.LocalKey
		parentAppKey, _ = strings.CutSuffix(parentAppKey, "_"+projecthelper.CalcProjectEnvKey(targetApp.ParentApp.Env))
		targetApp.LocalKey = parentAppKey + "_" + targetApp.LocalKey
	}
	if targetApp.Env != "" {
		targetApp.LocalKey += "_" + projecthelper.CalcProjectEnvKey(targetApp.Env)
	}
	targetApp.Key = data.TargetProject.Key + "_" + targetApp.LocalKey

	// App keys must be unique globally
	conflictApp, err := s.appRepo.GetByKey(ctx, db, "", targetApp.Key, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.New(err)
	}
	if conflictApp != nil {
		return apperrors.NewAlreadyExist("App").
			WithMsgLog("app key '%s' already exists", targetApp.Key)
	}

	// Create local network for the app to attach
	_, err = s.networkService.GetOrCreateProjectNetwork(ctx, data.TargetProject, targetApp.Env)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}

func (s *service) copyAppSettings(
	ctx context.Context,
	db database.IDB,
	data *appCopyData,
) (err error) {
	appSettings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.scope = ?", base.ObjectScopeApp),
		bunex.SelectWhere("setting.object_id = ?", data.SrcApp.ID),
	)
	if err != nil {
		return apperrors.New(err)
	}

	targetApp := data.TargetApp
	for _, setting := range appSettings {
		cpSetting, err := setting.Copy(true)
		if err != nil {
			return apperrors.New(err)
		}
		cpSetting.ObjectID = targetApp.ID
		cpSetting.CreatedAt = data.TimeNow
		cpSetting.UpdatedAt = data.TimeNow
		cpSetting.UpdateVer = 0
		st, err := data.OnCopySetting(targetApp, cpSetting)
		if err != nil {
			return apperrors.New(err)
		}
		if st != nil {
			data.CopiedSettings = append(data.CopiedSettings, st)
		}
	}

	targetApp.Settings = data.CopiedSettings

	// Update ref app for every sched job
	for _, jobSetting := range targetApp.GetSettingsByType(base.SettingTypeSchedJob) {
		schedJob := jobSetting.MustAsSchedJob()
		schedJob.App.ID = targetApp.ID
		jobSetting.MustSetData(schedJob)
	}

	// Validation

	// Active domains of the app need to validate
	newHttpSetting := targetApp.GetSettingByType(base.SettingTypeAppHttp)
	if newHttpSetting != nil {
		activeDomains := newHttpSetting.MustAsAppHttpSettings().GetActiveDomainNames()

		// Verify domains are allowed in project
		err = s.domainService.VerifyProjectDomains(ctx, db, targetApp.ProjectID, activeDomains)
		if err != nil {
			return apperrors.New(err)
		}

		// Make sure all domains used by the app are not hold by any other app
		err = s.domainService.VerifyDomainsAvailable(ctx, db, activeDomains, []string{targetApp.ID})
		if err != nil {
			return apperrors.New(err)
		}
	}

	return nil
}

func (s *service) persistAppData(
	ctx context.Context,
	db database.IDB,
	data *appCopyData,
) (err error) {
	app := data.TargetApp
	err = s.appRepo.Upsert(ctx, db, app,
		entity.AppUpsertingConflictCols, entity.AppUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}

	err = s.settingRepo.UpsertMulti(ctx, db, data.CopiedSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}

	// Loads all ref objects of the settings
	data.RefObjects, err = s.settingService.LoadReferenceObjects(ctx, db, app.GetObjectScope(),
		true, true, app.Settings...)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}

func (s *service) cleanupOnFail(
	ctx context.Context,
	data *appCopyData,
	err error,
) error {
	if err == nil {
		return nil
	}
	// Remove all created objects in docker
	if data.TargetService != nil && data.TargetService.ID != "" {
		_ = s.clusterService.ServiceRemove(ctx, data.TargetService.ID, clusterservice.ItemRemovalRetryMax, 0)
	}

	var secretIDs []string
	for _, secret := range data.TargetSecrets {
		secretIDs = append(secretIDs, secret.SecretID)
	}
	_ = s.clusterService.SecretsRemove(ctx, secretIDs, clusterservice.ItemRemovalRetryMax, 0)

	var configIDs []string
	for _, cfg := range data.TargetConfig {
		configIDs = append(configIDs, cfg.ConfigID)
	}
	_ = s.clusterService.ConfigsRemove(ctx, configIDs, clusterservice.ItemRemovalRetryMax, 0)
	return nil
}
