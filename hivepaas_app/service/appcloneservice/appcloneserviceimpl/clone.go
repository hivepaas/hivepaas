package appcloneserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
	ClonedSettings []*entity.Setting

	TimeNow time.Time
}

func (s *service) CloneApp(
	ctx context.Context,
	db database.IDB,
	req *appcloneservice.AppCloneReq,
) (resp *appcloneservice.AppCloneResp, err error) {
	resp = &appcloneservice.AppCloneResp{}
	data := &appCloneData{
		AppCloneReq: req,
		TimeNow:     timeutil.NowUTC(),
	}
	err = s.loadAppCloneData(ctx, db, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.Join(err, hperrors.NewPanic(r))
		}
		_ = s.cleanupOnFail(ctx, data, err)
	}()

	// Cloning steps

	if data.OnCloneStart != nil {
		err = data.OnCloneStart(req)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	err = s.cloneApp(ctx, db, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = s.cloneAppSettings(ctx, db, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = s.cloneSwarmService(ctx, db, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = s.cloneVolumes(ctx, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = s.persistAppData(ctx, db, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Post cloning steps

	// The copy's environment, config files, secrets, routing and scheduled jobs
	// are what its settings describe, which provisioning applies the same way to
	// an app made from a template. What is left below is cloning's own: a clone
	// has a source app, and nothing created from nothing does.
	err = s.applyClonedConfiguration(ctx, db, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = s.applyFinalContainerSettings(ctx, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = s.runCommands(ctx, db, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
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
	if data.RefObjects == nil {
		data.RefObjects = entity.NewRefObjects()
	}

	if data.SrcApp == nil {
		taskArgs, err := data.Task.ArgsAsAppClone()
		if err != nil {
			return hperrors.Wrap(err)
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
			return hperrors.Wrap(err)
		}
		data.SrcApp = app
		data.DropGatedMounts = taskArgs.DropGatedMounts

		// The settings this task was scheduled for. They are what the app is
		// loaded with above, and nothing else reads them.
		//
		// Missing is an error rather than a clone of nothing: a task exists
		// because somebody saved these settings and asked for them to run, and
		// the name the request validated as available comes from them. Without
		// them the clone would quietly be a different app than the one asked for.
		if data.CloneSettings == nil {
			cloneSetting := app.GetSettingByType(base.SettingTypeAppClone)
			if cloneSetting == nil {
				return hperrors.NewNotFound("App clone settings")
			}
			if data.CloneSettings, err = cloneSetting.AsAppCloneSettings(); err != nil {
				return hperrors.Wrap(err)
			}
		}
	}

	// A caller that brought its own source app decides for itself what to clone;
	// creating a preview app is the one that does.
	if data.CloneSettings == nil {
		data.CloneSettings = &entity.AppCloneSettings{}
	}

	// Loads all ref objects of the settings
	err = s.settingService.LoadRefObjectsByIDs(ctx, db, &data.RefObjects, data.SrcApp.GetObjectScope(),
		true, data.CloneSettings.GetRefObjectIDs())
	if err != nil {
		return hperrors.Wrap(err)
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
		return hperrors.Wrap(err)
	}

	err = s.settingRepo.UpsertMulti(ctx, db, data.ClonedSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Loads all ref objects of the settings
	err = s.settingService.LoadRefObjectsSkipMissing(ctx, db, &data.RefObjects, destApp.GetObjectScope(),
		false, destApp.Settings...)
	if err != nil {
		return hperrors.Wrap(err)
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

	// And its setting mounts' objects, which were made with the service.
	if data.DestApp != nil && data.DestApp.ID != "" {
		_ = s.settingMountService.RemoveApp(ctx, data.DestApp.ID)
	}
	return nil
}
