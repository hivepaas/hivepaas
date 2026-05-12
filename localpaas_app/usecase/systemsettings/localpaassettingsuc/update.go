package localpaassettingsuc

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/localpaassettingsuc/localpaassettingsdto"
)

const (
	serviceUpdateMaxRetry      = 2
	serviceUpdateRetryInterval = time.Second * 3
)

func (uc *UC) UpdateLocalPaaSSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *localpaassettingsdto.UpdateLocalPaaSSettingsReq,
) (*localpaassettingsdto.UpdateLocalPaaSSettingsResp, error) {
	req.Type = currentSettingType
	newSettings := req.ToEntity()

	var mainAppSvc, workerSvc *swarm.Service
	restartWorkers := false
	restartMainApp := false
	taskQueueStopped := false

	_, err := uc.UpdateUniqueSetting(ctx, &req.UpdateUniqueSettingReq, &settings.UpdateUniqueSettingData{
		Name: string(currentSettingType),
		AfterLoading: func(ctx context.Context, db database.Tx, data *settings.UpdateUniqueSettingData) error {
			lpSettings, err := data.Setting.AsLocalPaaSSettings()
			if err != nil {
				return apperrors.Wrap(err)
			}

			mainAppSvc, err = uc.lpAppService.GetLpAppSwarmService(ctx)
			if err != nil {
				return apperrors.Wrap(err)
			}
			workerSvc, err = uc.lpAppService.GetLpWorkerSwarmService(ctx)
			if err != nil {
				return apperrors.Wrap(err)
			}

			if newSettings.WorkerSettings.Replicas < lpSettings.WorkerSettings.Replicas {
				restartWorkers = true
			}
			if newSettings.WorkerSettings.RunWorkerInMainApp != lpSettings.WorkerSettings.RunWorkerInMainApp {
				restartMainApp = true
			}
			if newSettings.WorkerSettings.Concurrency != lpSettings.WorkerSettings.Concurrency ||
				newSettings.TaskSettings.TaskCheckInterval != lpSettings.TaskSettings.TaskCheckInterval ||
				newSettings.TaskSettings.TaskCreateInterval != lpSettings.TaskSettings.TaskCreateInterval {
				restartWorkers = true
				restartMainApp = true
			}

			if restartWorkers || restartMainApp {
				// Make sure there is no task in-progress
				_, err = uc.taskService.LockAllPendingTasks(ctx, db, time.Second*10) //nolint:mnd
				if err != nil {
					return apperrors.New(err)
				}
				// Stop all workers from executing new jobs
				err = uc.taskQueue.StopAllSchedulers()
				if err != nil {
					return apperrors.New(err)
				}
				taskQueueStopped = true
			}

			return nil
		},
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			err := pData.Setting.SetData(newSettings)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if restartWorkers {
		err = gofn.ExecRetry(func() error {
			return uc.applySettingsToWorkerService(ctx, mainAppSvc, workerSvc, newSettings)
		}, serviceUpdateMaxRetry, serviceUpdateRetryInterval)
	}

	if restartMainApp {
		err2 := gofn.ExecRetry(func() error {
			return uc.applySettingsToMainService(ctx, mainAppSvc, newSettings)
		}, serviceUpdateMaxRetry, serviceUpdateRetryInterval)
		err = errors.Join(err, err2)
	}

	// When task queue(s) are shutdown but the application fails
	if err != nil && taskQueueStopped {
		_ = uc.taskQueue.StartAllSchedulers()
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &localpaassettingsdto.UpdateLocalPaaSSettingsResp{}, nil
}

func (uc *UC) applySettingsToMainService(
	ctx context.Context,
	mainAppSvc *swarm.Service,
	_ *entity.LocalPaaSSettings,
) error {
	// Just restart it for now
	mainAppSvc.Spec.TaskTemplate.ForceUpdate++
	_, err := uc.dockerManager.ServiceUpdate(ctx, mainAppSvc.ID, &mainAppSvc.Version, &mainAppSvc.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (uc *UC) applySettingsToWorkerService(
	ctx context.Context,
	mainAppSvc, workerSvc *swarm.Service,
	lpSettings *entity.LocalPaaSSettings,
) error {
	uc.lpAppService.SyncLpWorkerSwarmServiceConfig(mainAppSvc, workerSvc)

	// Set service mode and replicas
	workerSvc.Spec.Mode.Replicated = &swarm.ReplicatedService{
		Replicas: new(uint64(lpSettings.WorkerSettings.Replicas)), //nolint:gosec
	}

	_, err := uc.dockerManager.ServiceUpdate(ctx, workerSvc.ID, &workerSvc.Version, &workerSvc.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
