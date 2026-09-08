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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingsrevertservice"
)

// revertTraefikConfig puts traefik's startup command back.
//
// This is the trial for the failure swarm cannot see. A command that traefik
// refuses to parse makes the process exit, and the healthcheck plus
// failure_action: rollback on the service already answer that. What they cannot
// answer is a command traefik accepts and serves nothing under - a provider
// constraint that matches no service, a default middleware that does not resolve
// - because /ping keeps returning 200 the whole time. Nothing but somebody
// failing to come back tells us about that.
//
// The undo runs from the HivePaaS app process over the docker socket, so it does
// not need traefik to be working in order to fix traefik.
func (s *service) revertTraefikConfig(
	ctx context.Context,
	db database.Tx,
	args *entity.TaskSettingsRevertArgs,
) (*settingsrevertservice.RevertResp, error) {
	// By id, not by type. The task records exactly which row it put on trial, and
	// asking for "the traefik config setting" instead would let a revert land on a
	// row this trial was never about.
	setting, err := s.settingRepo.GetByID(ctx, db, nil, base.SettingTypeTraefikConfig, args.SettingID, true,
		bunex.SelectFor("UPDATE"),
	)
	// GetSingle reports a missing row as ErrNotFound rather than a nil setting, so
	// both shapes are checked. Either one means the row this trial was about is
	// gone, which is a reason to stop rather than a failure to retry: there is
	// nothing left to restore it onto.
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if setting == nil {
		return &settingsrevertservice.RevertResp{Reason: "setting no longer exists"}, nil
	}
	if setting.UpdateVer != args.ProbationVer {
		return &settingsrevertservice.RevertResp{Reason: "settings changed since"}, nil
	}

	args.Snapshot.RestoreTo(setting)
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()

	restored, err := setting.AsTraefikConfig()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if err = s.settingRepo.Update(ctx, db, setting); err != nil {
		return nil, hperrors.Wrap(err)
	}

	if err = s.restoreTraefikCommand(ctx, restored); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &settingsrevertservice.RevertResp{Reverted: true}, nil
}

// restoreTraefikCommand pushes the restored command onto the swarm service,
// unless it is already what is running.
//
// Swarm may have got there first. A command traefik cannot start under fails its
// healthcheck, and the service's own failure_action: rollback then restores the
// previous spec - which is this same command. Pushing it again would recreate
// traefik's task for no change at all, taking ingress down a second time on the
// way out of an outage. So the live args are compared first, and an equal spec is
// left alone.
func (s *service) restoreTraefikCommand(ctx context.Context, restored *entity.TraefikConfig) error {
	traefikSvc, err := s.traefikService.GetTraefikSwarmService(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if traefikSvc.Spec.TaskTemplate.ContainerSpec == nil {
		traefikSvc.Spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
	}
	if restored.SameArgsAs(traefikSvc.Spec.TaskTemplate.ContainerSpec.Args) {
		s.logger.Infof("traefik is already running the command being restored, leaving its task alone")
		return nil
	}

	traefikSvc.Spec.TaskTemplate.ContainerSpec.Args = restored.Args
	setRollbackOnFailure(&traefikSvc.Spec)

	if _, err = s.dockerManager.ServiceUpdate(ctx, traefikSvc.ID,
		&traefikSvc.Version, &traefikSvc.Spec); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
