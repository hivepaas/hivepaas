package appcloneserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
)

type appCloneData struct {
	*appcloneservice.AppCloneReq

	DestApp        *entity.App
	SrcService     *swarm.Service
	DestService    *swarm.Service
	DestSecrets    []*entity.SwarmSecretRef
	DestConfig     []*entity.SwarmConfigRef
	ClonedSettings []*entity.Setting

	TimeNow time.Time
}

func (s *service) CloneApp(
	ctx context.Context,
	db database.Tx,
	req *appcloneservice.AppCloneReq,
) (resp *appcloneservice.AppCloneResp, err error) {
	resp = &appcloneservice.AppCloneResp{}
	data := &appCloneData{
		AppCloneReq: req,
		TimeNow:     timeutil.NowUTC(),
	}
	err = s.loadAppCloneData(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.Join(err, apperrors.NewPanic(r))
		}
		_ = s.cleanupOnFail(ctx, data, err)
	}()

	// Cloning steps

	err = s.cloneApp(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.cloneAppSettings(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.cloneSwarmService(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.cloneVolumes(ctx, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.persistAppData(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Post cloning steps

	err = s.applyEnvVars(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.applySwarmConfigFiles(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.applySwarmSecrets(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.applyAppHttpSettings(ctx, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.applySchedJobSettings(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.applyFinalContainerSettings(ctx, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.runCommands(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.TargetApp = data.DestApp
	resp.TargetService = data.DestService
	resp.OnCleanup = func(e error) error {
		return s.cleanupOnFail(ctx, data, e)
	}
	return resp, nil
}

func (s *service) loadAppCloneData(
	ctx context.Context,
	db database.IDB,
	data *appCloneData,
) (err error) {
	if data.SrcApp != nil && data.ClonedSettings != nil {
		return nil
	}

	taskArgs, err := data.Task.ArgsAsAppClone()
	if err != nil {
		return apperrors.Wrap(err)
	}

	app, err := s.appService.LoadApp(ctx, db, "", taskArgs.SrcApp.ID, true, true,
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("Project.ProjectEnvs"),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppClone),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.SrcApp = app

	cloneSetting := app.GetSettingByType(base.SettingTypeAppClone)
	if cloneSetting == nil {
		return apperrors.NewNotFound("App clone settings")
	}
	cloneSettings := cloneSetting.MustAsAppCloneSettings()
	data.CloneSettings = cloneSettings

	// Loads all ref objects of the settings
	data.RefObjects, err = s.settingService.LoadReferenceObjects(ctx, db, app.GetObjectScope(),
		true, true, cloneSetting)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) persistAppData(
	ctx context.Context,
	db database.IDB,
	data *appCloneData,
) (err error) {
	destApp := data.DestApp
	err = s.appRepo.Upsert(ctx, db, destApp,
		entity.AppUpsertingConflictCols, entity.AppUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = s.settingRepo.UpsertMulti(ctx, db, data.ClonedSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Loads all ref objects of the settings
	// TODO: review usage of requireActive=false && errorIfUnavail=false
	data.RefObjects, err = s.settingService.LoadReferenceObjects(ctx, db, destApp.GetObjectScope(),
		false, false, destApp.Settings...)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) cleanupOnFail(
	ctx context.Context,
	data *appCloneData,
	err error,
) error {
	if err == nil {
		return nil
	}
	// Remove all created objects in docker
	if data.DestService != nil && data.DestService.ID != "" {
		_ = s.clusterService.ServiceRemove(ctx, data.DestService.ID, clusterservice.ItemRemovalRetryMax, 0)
	}

	var secretIDs []string
	for _, secret := range data.DestSecrets {
		secretIDs = append(secretIDs, secret.SecretID)
	}
	_ = s.clusterService.SecretsRemove(ctx, secretIDs, clusterservice.ItemRemovalRetryMax, 0)

	var configIDs []string
	for _, cfg := range data.DestConfig {
		configIDs = append(configIDs, cfg.ConfigID)
	}
	_ = s.clusterService.ConfigsRemove(ctx, configIDs, clusterservice.ItemRemovalRetryMax, 0)
	return nil
}
