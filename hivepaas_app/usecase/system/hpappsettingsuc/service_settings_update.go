package hpappsettingsuc

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

const (
	serviceUpdateMaxRetry      = 2
	serviceUpdateRetryInterval = time.Second * 3
)

func (uc *UC) UpdateServiceSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.UpdateServiceSettingsReq,
) (*hpappsettingsdto.UpdateServiceSettingsResp, error) {
	var data *updateServiceSettingsData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data = &updateServiceSettingsData{}
		err := uc.loadServiceSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.New(err)
		}

		persistingData := &persistingSettingsData{}
		uc.prepareUpdatingServiceSettings(data, persistingData)

		err = uc.persistSettingsData(ctx, db, persistingData)
		if err != nil {
			return apperrors.New(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	if data.workerSvcChanges {
		e := gofn.ExecRetry(func() error {
			return uc.applyServiceSettingsToWorkerService(ctx, data)
		}, serviceUpdateMaxRetry, serviceUpdateRetryInterval)

		// When task queue(s) are shutdown but the setting application fails,
		// restart the services to make sure they run properly.
		if e != nil && data.taskQueueStopped {
			_ = gofn.ExecRetry(func() error {
				return uc.hpAppService.RestartHpWorkerSwarmService(ctx)
			}, serviceUpdateMaxRetry, serviceUpdateRetryInterval)
		}
		err = errors.Join(err, e)
	}

	if data.mainSvcChanges {
		e := gofn.ExecRetry(func() error {
			return uc.applyServiceSettingsToMainService(ctx, data)
		}, serviceUpdateMaxRetry, serviceUpdateRetryInterval)

		// When task queue(s) are shutdown but the setting application fails,
		// restart the services to make sure they run properly.
		if e != nil && data.taskQueueStopped {
			_ = gofn.ExecRetry(func() error {
				return uc.hpAppService.RestartHpAppSwarmService(ctx)
			}, serviceUpdateMaxRetry, serviceUpdateRetryInterval)
		}
		err = errors.Join(err, e)
	}

	if err != nil {
		return nil, apperrors.New(err)
	}

	return &hpappsettingsdto.UpdateServiceSettingsResp{}, nil
}

type updateServiceSettingsData struct {
	Setting       *entity.Setting
	NewSettings   *entity.HivePaaSService
	MainService   *swarm.Service
	WorkerService *swarm.Service

	workerSvcChanges bool
	mainSvcChanges   bool
	taskQueueStopped bool
}

func (uc *UC) loadServiceSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *hpappsettingsdto.UpdateServiceSettingsReq,
	data *updateServiceSettingsData,
) error {
	setting, err := uc.settingRepo.GetSingle(ctx, db, nil, base.SettingTypeHivePaaSService, true,
		bunex.SelectFor("UPDATE"),
	)
	if err != nil {
		return apperrors.New(err)
	}
	data.Setting = setting

	if setting != nil && setting.UpdateVer != req.UpdateVer {
		return apperrors.New(apperrors.ErrUpdateVerMismatched)
	}

	newSettings := req.ToEntity()
	data.NewSettings = newSettings

	currSettings, err := data.Setting.AsHivePaaSService()
	if err != nil {
		return apperrors.New(err)
	}

	mainAppSvc, err := uc.hpAppService.GetHpAppSwarmService(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	data.MainService = mainAppSvc

	workerSvc, err := uc.hpAppService.GetHpWorkerSwarmService(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	data.WorkerService = workerSvc

	if newSettings.AppSettings.Replicas != currSettings.AppSettings.Replicas ||
		newSettings.WorkerSettings.RunWorkerInMainApp != currSettings.WorkerSettings.RunWorkerInMainApp {
		data.mainSvcChanges = true
	}
	if newSettings.WorkerSettings.Replicas != currSettings.WorkerSettings.Replicas {
		data.workerSvcChanges = true
	}
	if newSettings.WorkerSettings.Concurrency != currSettings.WorkerSettings.Concurrency ||
		newSettings.TaskSettings.TaskCheckInterval != currSettings.TaskSettings.TaskCheckInterval ||
		newSettings.TaskSettings.TaskCreateInterval != currSettings.TaskSettings.TaskCreateInterval ||
		newSettings.HealthcheckSettings.BaseInterval != currSettings.HealthcheckSettings.BaseInterval {
		data.workerSvcChanges = true
		data.mainSvcChanges = true
	}

	if data.workerSvcChanges || data.mainSvcChanges {
		// Make sure there is no task in-progress
		_, err = uc.taskService.LockAllPendingTasks(ctx, db, time.Second*10) //nolint:mnd
		if err != nil {
			return apperrors.New(err)
		}
		// Stop all workers from taking new jobs
		err = uc.taskQueue.StopAllSchedulers()
		if err != nil {
			return apperrors.New(err)
		}
		data.taskQueueStopped = true
	}

	return nil
}

func (uc *UC) applyServiceSettingsToMainService(
	ctx context.Context,
	data *updateServiceSettingsData,
) error {
	mainAppSvc := data.MainService

	// Set service mode and replicas
	mainAppSvc.Spec.Mode.Replicated = &swarm.ReplicatedService{
		Replicas: new(uint64(data.NewSettings.AppSettings.Replicas)), //nolint:gosec
	}
	mainAppSvc.Spec.TaskTemplate.ForceUpdate++

	_, err := uc.dockerManager.ServiceUpdate(ctx, mainAppSvc.ID, &mainAppSvc.Version, &mainAppSvc.Spec)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (uc *UC) applyServiceSettingsToWorkerService(
	ctx context.Context,
	data *updateServiceSettingsData,
) error {
	mainAppSvc, workerSvc := data.MainService, data.WorkerService
	uc.hpAppService.SyncHpWorkerSwarmServiceConfig(mainAppSvc, workerSvc)

	// Set service mode and replicas
	workerSvc.Spec.Mode.Replicated = &swarm.ReplicatedService{
		Replicas: new(uint64(data.NewSettings.WorkerSettings.Replicas)), //nolint:gosec
	}
	workerSvc.Spec.TaskTemplate.ForceUpdate++

	_, err := uc.dockerManager.ServiceUpdate(ctx, workerSvc.ID, &workerSvc.Version, &workerSvc.Spec)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (uc *UC) prepareUpdatingServiceSettings(
	data *updateServiceSettingsData,
	persistingData *persistingSettingsData,
) {
	setting := data.Setting
	setting.MustSetData(data.NewSettings)
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()

	persistingData.Settings = append(persistingData.Settings, setting)
}

type persistingSettingsData struct {
	Settings []*entity.Setting
}

func (uc *UC) persistSettingsData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingSettingsData,
) error {
	err := uc.settingRepo.UpsertMulti(ctx, db, persistingData.Settings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
