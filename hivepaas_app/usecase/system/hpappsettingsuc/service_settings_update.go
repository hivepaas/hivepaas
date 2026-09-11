package hpappsettingsuc

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
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
			return hperrors.Wrap(err)
		}

		persistingData := &persistingSettingsData{}
		uc.prepareUpdatingServiceSettings(data, persistingData)

		if data.proxyChanges {
			// Same transaction as the change, so a committed change always has a
			// committed deadline.
			err = uc.armProbation(ctx, db, auth,
				&probationArgs{
					AppID:       data.AppID,
					Setting:     data.Setting,
					Snapshot:    data.Snapshot,
					Window:      data.ProbationWindow,
					SettleDelay: data.SettleDelay,
				},
				&data.probationResult,
				func(task *entity.Task) {
					persistingData.Tasks = append(persistingData.Tasks, task)
				},
			)
			if err != nil {
				return hperrors.Wrap(err)
			}
		}

		err = uc.persistSettingsData(ctx, db, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return uc.recordHivePaaSSettingsUpdate(ctx, db, auth, "service", auditdetail.New().
			Set("proxyChanged", data.proxyChanges).
			Set("onProbation", data.probationResult.Probation != nil).
			WithChangedFields(
				settingsSnapshotData(base.SettingTypeHivePaaSService, data.Snapshot),
				data.NewSettings))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if data.workerSvcChanges {
		// No outer retry: applyServiceSettingsToWorkerService retries internally,
		// re-inspecting the service each time. Wrapping it in another retry only
		// multiplied the attempts.
		e := uc.applyServiceSettingsToWorkerService(ctx, data)

		// When task queue(s) are shutdown but the setting application fails,
		// restart the services to make sure they run properly.
		if e != nil && data.taskQueueStopped {
			_ = gofn.ExecRetry(func() error {
				return uc.hpAppService.RestartHpWorkerSwarmService(ctx)
			}, serviceUpdateMaxRetry, serviceUpdateRetryInterval)
		}
		err = errors.Join(err, e)
	}

	// Apply the trusted IPs to traefik, including when there are none.
	//
	// Withdrawing the proxy has to reach Traefik as much as declaring one does.
	// Skipping the call when no provider is set - which is what this used to do -
	// left the entrypoint trusting a proxy that is no longer in front, so
	// X-Forwarded-For stayed honored from whoever connected. Validation clears the
	// addresses when the provider is cleared, so the empty list here is the
	// instruction to stop trusting.
	_, e := uc.traefikService.ApplyTrustedIPsToWebEntrypoints(ctx, &traefikservice.ApplyTrustedIPsReq{
		TrustedIPs: req.ProxySettings.TrustedIPs,
	})
	err = errors.Join(err, e)

	// The proxy topology is baked into each app's labels, not read per request, so
	// without this the change has no effect at all - and the trial armed above
	// would be guarding something that had not happened. It runs after the traefik
	// arguments, because whether a depth is written at all depends on the
	// entrypoint trusting a proxy, and that is what those arguments say.
	err = errors.Join(err, uc.reapplyClientIPStrategy(ctx, data))

	// Once traefik has the new arguments and the apps have the new labels, the
	// change is real and the deadline can start counting.
	uc.scheduleProbation(ctx, &data.probationResult)

	// Last, because this one recreates the task running this code.
	//
	// Swarm answers a TaskTemplate change by stopping the current task, so from
	// here on everything is racing a shutdown. Graceful shutdown does let an
	// in-flight request finish - fx stops the HTTP server before the database and
	// http.Server.Shutdown waits - but nothing about those two facts is written
	// down as load-bearing for this function, and neither is gin leaving the
	// request context uncancellable. Putting the self-restart at the end means
	// none of that has to hold.
	if data.mainSvcChanges {
		// See applyServiceSettingsToWorkerService on the missing outer retry.
		e := uc.applyServiceSettingsToMainService(ctx, data)

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
		// The caller gets an error and no pendingChange, so nothing tells them a
		// trial is running - and in a few minutes it will undo their change with
		// no explanation. Reverting is the right outcome for a half-applied
		// change; being surprised by it is not, so at least say so here.
		if data.Probation != nil {
			uc.logger.Warnf("service settings apply failed with a change on trial: probation %s "+
				"will revert it at %v unless it is confirmed", data.Probation.ID, data.Probation.RunAt)
		}
		return nil, hperrors.Wrap(err)
	}

	return &hpappsettingsdto.UpdateServiceSettingsResp{
		Data: hpappsettingsdto.TransformPendingChange(data.Probation),
	}, nil
}

type updateServiceSettingsData struct {
	Setting       *entity.Setting
	NewSettings   *entity.HivePaaSService
	MainService   *swarm.Service
	WorkerService *swarm.Service
	AppID         string

	// Snapshot is the settings as they were before this request, and the state a
	// revert restores. Taken before MustSetData overwrites the payload.
	Snapshot        entity.SettingSnapshot
	SettleDelay     time.Duration
	ProbationWindow time.Duration
	proxyChanges    bool

	probationResult

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
		return hperrors.Wrap(err)
	}
	data.Setting = setting

	if setting != nil && setting.UpdateVer != req.UpdateVer {
		return hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}

	data.Snapshot = entity.SettingSnapshotOf(setting)

	newSettings := req.ToEntity()
	data.NewSettings = newSettings

	currSettings, err := data.Setting.AsHivePaaSService()
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Only the proxy fields decide how a client address is read, and therefore
	// whether the IP allowlists on the HivePaaS routers still admit the caller.
	// They are also the only ones traefik has no healthcheck-driven rollback for -
	// see the comment on that healthcheck. Everything else in here either cannot
	// lock anybody out or is already covered by swarm rolling the service back.
	data.proxyChanges = !newSettings.ProxySettings.Equal(&currSettings.ProxySettings)

	app, err := uc.hpAppService.LoadAppByKey(ctx, db, base.HivepaasAppKey,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.AppID = app.ID

	mainAppSvc, err := uc.hpAppService.GetHpAppSwarmService(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.MainService = mainAppSvc

	workerSvc, err := uc.hpAppService.GetHpWorkerSwarmService(ctx)
	if err != nil {
		return hperrors.Wrap(err)
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
		newSettings.PeriodicSettings.BaseInterval != currSettings.PeriodicSettings.BaseInterval {
		data.workerSvcChanges = true
		data.mainSvcChanges = true
	}
	if newSettings.PeriodicSettings.BaseInterval != currSettings.PeriodicSettings.BaseInterval {
		data.workerSvcChanges = true
		if currSettings.WorkerSettings.RunWorkerInMainApp {
			data.mainSvcChanges = true
		}
	}

	// Last, because it depends on mainSvcChanges, which is only settled above.
	//
	// A proxy-only change leaves the app running and is back in seconds; the same
	// request carrying a replica or worker setting takes the app down with it, and
	// until it is serving again the dashboard must not read a failed probe as a
	// lockout. See SettingsProbationAppRestartSettleDelay.
	data.SettleDelay = entity.SettingsProbationSettleDelay
	if data.mainSvcChanges {
		data.SettleDelay = entity.SettingsProbationAppRestartSettleDelay
	}
	data.ProbationWindow = resolveProbationWindow(req.ConfirmWindow.ToDuration(), data.SettleDelay)

	if data.workerSvcChanges || data.mainSvcChanges {
		// Make sure there is no task in-progress
		_, err = uc.taskService.LockAllPendingTasks(ctx, db, time.Second*10) //nolint:mnd
		if err != nil {
			return hperrors.Wrap(err)
		}
		// Stop all workers from taking new jobs
		err = uc.taskQueue.StopAllSchedulers()
		if err != nil {
			return hperrors.Wrap(err)
		}
		data.taskQueueStopped = true
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
	Tasks    []*entity.Task
}

func (uc *UC) persistSettingsData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingSettingsData,
) error {
	err := uc.settingRepo.UpsertMulti(ctx, db, persistingData.Settings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}
	err = uc.taskRepo.UpsertMulti(ctx, db, persistingData.Tasks,
		entity.TaskUpsertingConflictCols, entity.TaskUpsertingUpdateCols)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// reapplyClientIPStrategy pushes the new topology onto the HivePaaS app.
//
// Only that one, and only when the proxy settings actually moved. The change has
// to be real for the app the trial is about, or the countdown would be guarding
// something that had not happened; the other apps are swept once somebody
// confirms - see scheduleAppLabelsSweep.
func (uc *UC) reapplyClientIPStrategy(ctx context.Context, data *updateServiceSettingsData) error {
	if !data.proxyChanges {
		return nil
	}

	var resp *approutingservice.ReapplyClientIPStrategyResp
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		var e error
		resp, e = uc.appRoutingService.ReapplyClientIPStrategy(ctx, db,
			&approutingservice.ReapplyClientIPStrategyReq{
				PrimaryAppID: data.AppID,
				PrimaryOnly:  true,
			})
		return hperrors.Wrap(e)
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Reported, not fatal. The apps that did get through are correct, HivePaaS is
	// among them or this would have returned an error, and the trial still holds
	// the undo for all of it.
	for appID, reason := range resp.Failed {
		uc.logger.Errorf("failed to reapply the client IP strategy to app %s: %s", appID, reason)
	}

	return nil
}

// applyServiceSettingsToMainService pushes the replica count onto the HivePaaS
// main service.
//
// The service is re-inspected instead of reusing the copy loaded at the start of
// the request, because by the time this runs that copy is two ways stale. A proxy
// change in the same request has already gone through reapplyClientIPStrategy,
// which does its own ServiceUpdate on this very service: its version index has
// moved on - swarm answers a stale one with "update out of sequence", and
// retrying the same stale index just fails again - and its spec now carries the
// regenerated ip-strategy labels that pushing the old spec would silently undo.
// ServiceUpdateFunc with a nil service inspects before every attempt, so both the
// version and the labels come from whatever is live now.
func (uc *UC) applyServiceSettingsToMainService(
	ctx context.Context,
	data *updateServiceSettingsData,
) error {
	err := uc.dockerManager.ServiceUpdateFunc(ctx, data.MainService.ID, nil,
		func(_ int, mainAppSvc *swarm.Service) (bool, error) {
			data.MainService = mainAppSvc

			// Set service mode and replicas
			mainAppSvc.Spec.Mode.Replicated = &swarm.ReplicatedService{
				Replicas: new(uint64(data.NewSettings.AppSettings.Replicas)), //nolint:gosec
			}
			mainAppSvc.Spec.TaskTemplate.ForceUpdate++

			if mainAppSvc.Spec.UpdateConfig == nil {
				mainAppSvc.Spec.UpdateConfig = &swarm.UpdateConfig{}
			}
			mainAppSvc.Spec.UpdateConfig.FailureAction = swarm.UpdateFailureActionRollback
			mainAppSvc.Spec.UpdateConfig.MaxFailureRatio = 0.5

			return true, nil
		}, serviceUpdateMaxRetry, serviceUpdateRetryInterval)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// applyServiceSettingsToWorkerService pushes the replica count onto the HivePaaS
// worker service.
//
// Re-inspected per attempt for the same reason as the main service: a retry that
// replays the version index that just lost is guaranteed to lose again. Nothing
// earlier in this request touches the worker service, so the first attempt is only
// paying an extra inspect - the retries are what this buys.
func (uc *UC) applyServiceSettingsToWorkerService(
	ctx context.Context,
	data *updateServiceSettingsData,
) error {
	err := uc.dockerManager.ServiceUpdateFunc(ctx, data.WorkerService.ID, nil,
		func(_ int, workerSvc *swarm.Service) (bool, error) {
			data.WorkerService = workerSvc
			uc.hpAppService.SyncHpWorkerSwarmServiceConfig(data.MainService, workerSvc)

			// Set service mode and replicas
			workerSvc.Spec.Mode.Replicated = &swarm.ReplicatedService{
				Replicas: new(uint64(data.NewSettings.WorkerSettings.Replicas)), //nolint:gosec
			}
			workerSvc.Spec.TaskTemplate.ForceUpdate++

			if workerSvc.Spec.UpdateConfig == nil {
				workerSvc.Spec.UpdateConfig = &swarm.UpdateConfig{}
			}
			workerSvc.Spec.UpdateConfig.FailureAction = swarm.UpdateFailureActionRollback
			workerSvc.Spec.UpdateConfig.MaxFailureRatio = 0.5

			return true, nil
		}, serviceUpdateMaxRetry, serviceUpdateRetryInterval)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
