package settingsrevertserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingsrevertservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

// updateMaxFailureRatio matches what every other swarm update in the codebase
// sets, so a revert converges on the same terms as the change it undoes.
const updateMaxFailureRatio = 0.5

// revertHivePaaSService puts the HivePaaS service settings back.
//
// What has to be re-applied is decided by comparing the snapshot against what is
// live, not by assuming the change was proxy-only. A trial is armed whenever the
// proxy settings move, and the same request may have carried a replica count with
// it; restoring the row without pushing those back would leave the database and
// the swarm services disagreeing, with nothing left to notice.
func (s *service) revertHivePaaSService(
	ctx context.Context,
	db database.Tx,
	args *entity.TaskSettingsRevertArgs,
) (*settingsrevertservice.RevertResp, error) {
	setting, err := s.settingRepo.GetSingle(ctx, db, nil, base.SettingTypeHivePaaSService, true,
		bunex.SelectFor("UPDATE"),
	)
	// GetSingle reports a missing row as ErrNotFound rather than a nil setting, so
	// both shapes are checked. Either one means the row this trial was about is
	// gone, which is a reason to stop rather than a failure to retry.
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if setting == nil {
		return &settingsrevertservice.RevertResp{Reason: "setting no longer exists"}, nil
	}
	if setting.UpdateVer != args.ProbationVer {
		return &settingsrevertservice.RevertResp{Reason: "settings changed since"}, nil
	}

	// Read the live settings before the snapshot overwrites them: they are the
	// other half of every comparison below.
	liveSettings, err := setting.AsHivePaaSService()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	args.Snapshot.RestoreTo(setting)
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()

	restored, err := setting.AsHivePaaSService()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = s.settingRepo.Update(ctx, db, setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if err = s.revertSwarmServices(ctx, liveSettings, restored); err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Always, including when the restored list is empty. An empty list is the
	// instruction to stop trusting a proxy, and skipping the call would leave
	// traefik honoring X-Forwarded-For from whoever connects - see
	// ApplyTrustedIPsToWebEntrypoints.
	_, err = s.traefikService.ApplyTrustedIPsToWebEntrypoints(ctx, &traefikservice.ApplyTrustedIPsReq{
		TrustedIPs: restored.ProxySettings.TrustedIPs,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// The mirror of the sweep the change itself ran, and just as narrow: only the
	// HivePaaS app was given the new depth while the change was on trial, so only
	// it has anything to put back. Nothing else was touched, which is the point of
	// deferring the rest until a change is confirmed.
	if err = s.reapplyClientIPStrategy(ctx, db, args.AppID); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &settingsrevertservice.RevertResp{Reverted: true}, nil
}

// revertSwarmServices pushes the restored settings back onto the service specs,
// but only where the restored value differs from the live one.
//
// Each of these recreates the task it touches. Doing them unconditionally would
// restart HivePaaS on every proxy revert, which is both the slowest way back and
// a second chance for something to go wrong.
func (s *service) revertSwarmServices(
	ctx context.Context,
	live, restored *entity.HivePaaSService,
) error {
	mainChanged := restored.AppSettings.Replicas != live.AppSettings.Replicas ||
		restored.WorkerSettings.RunWorkerInMainApp != live.WorkerSettings.RunWorkerInMainApp
	workerChanged := restored.WorkerSettings.Replicas != live.WorkerSettings.Replicas
	if restored.WorkerSettings.Concurrency != live.WorkerSettings.Concurrency ||
		restored.TaskSettings.TaskCheckInterval != live.TaskSettings.TaskCheckInterval ||
		restored.TaskSettings.TaskCreateInterval != live.TaskSettings.TaskCreateInterval ||
		restored.PeriodicSettings.BaseInterval != live.PeriodicSettings.BaseInterval {
		mainChanged = true
		workerChanged = true
	}
	if !mainChanged && !workerChanged {
		return nil
	}

	mainAppSvc, err := s.hpAppService.GetHpAppSwarmService(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if mainChanged {
		mainAppSvc.Spec.Mode.Replicated = &swarm.ReplicatedService{
			Replicas: new(uint64(restored.AppSettings.Replicas)), //nolint:gosec
		}
		mainAppSvc.Spec.TaskTemplate.ForceUpdate++
		setRollbackOnFailure(&mainAppSvc.Spec)

		if _, err = s.dockerManager.ServiceUpdate(ctx, mainAppSvc.ID, &mainAppSvc.Version, &mainAppSvc.Spec); err != nil {
			return hperrors.Wrap(err)
		}
	}

	if workerChanged {
		workerSvc, e := s.hpAppService.GetHpWorkerSwarmService(ctx)
		if e != nil {
			return hperrors.Wrap(e)
		}
		s.hpAppService.SyncHpWorkerSwarmServiceConfig(mainAppSvc, workerSvc)
		workerSvc.Spec.Mode.Replicated = &swarm.ReplicatedService{
			Replicas: new(uint64(restored.WorkerSettings.Replicas)), //nolint:gosec
		}
		workerSvc.Spec.TaskTemplate.ForceUpdate++
		setRollbackOnFailure(&workerSvc.Spec)

		if _, err = s.dockerManager.ServiceUpdate(ctx, workerSvc.ID, &workerSvc.Version, &workerSvc.Spec); err != nil {
			return hperrors.Wrap(err)
		}
	}

	return nil
}

func setRollbackOnFailure(spec *swarm.ServiceSpec) {
	if spec.UpdateConfig == nil {
		spec.UpdateConfig = &swarm.UpdateConfig{}
	}
	spec.UpdateConfig.FailureAction = swarm.UpdateFailureActionRollback
	spec.UpdateConfig.MaxFailureRatio = updateMaxFailureRatio
}

func (s *service) reapplyClientIPStrategy(ctx context.Context, db database.Tx, appID string) error {
	// Lenient, because this sweep is the way back. A HivePaaS app whose routing
	// references a setting deleted since would otherwise abort the revert - and
	// the primary app aborting is fatal to it - leaving the operator in the
	// configuration they are trying to escape.
	resp, err := s.appRoutingService.ReapplyClientIPStrategy(ctx, db,
		&approutingservice.ReapplyClientIPStrategyReq{
			PrimaryAppID:          appID,
			PrimaryOnly:           true,
			SkipMissingRefObjects: true,
		})
	if err != nil {
		return hperrors.Wrap(err)
	}
	for failedAppID, reason := range resp.Failed {
		s.logger.Errorf("failed to reapply the client IP strategy to app %s during a revert: %s",
			failedAppID, reason)
	}
	return nil
}
