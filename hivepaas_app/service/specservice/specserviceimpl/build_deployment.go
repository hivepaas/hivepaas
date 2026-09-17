package specserviceimpl

import (
	"context"
	"maps"
	"slices"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (s *service) buildSource(_ context.Context, state *buildState) error {
	block := specmodel.BlockDeploymentSource
	source := &entity.AppDeploymentSettings{}
	if err := decodeBlock(block, state.req.Doc.Deployment.Source, source); err != nil {
		return err
	}
	if source.ActiveMethod != base.DeploymentMethodImage || source.ImageSource == nil ||
		source.ImageSource.Image == "" {
		return invalidBlock(block, "an image to deploy is required")
	}
	return state.addSetting(base.SettingTypeAppDeployment, entity.CurrentAppDeploymentSettingsVersion, true, source)
}

// buildHealthcheck writes a healthcheck in the form docker runs it. CMD is an
// argv, so its command is split; CMD-SHELL is one string handed to the shell,
// so its command is not. A mode left empty means CMD-SHELL, which is what a
// template author writing a shell command expects.
//
// The container settings screen splits a CMD-SHELL command too, and docker then
// hands the shell only its first word: `sh -c pg_isready -U app` runs pg_isready
// with no arguments. That is a bug to fix there, not a behavior to copy here.
func (s *service) buildHealthcheck(_ context.Context, state *buildState) error {
	check := state.req.Doc.Deployment.Container.Healthcheck
	containerSpec := state.req.Spec.TaskTemplate.ContainerSpec
	if !check.Enabled {
		containerSpec.Healthcheck = nil
		return nil
	}

	var test []string
	switch check.Mode {
	case docker.HealthcheckModeCmd:
		command, err := executil.CmdSplit(check.Command)
		if err != nil {
			return invalidBlock(specmodel.BlockContainerHealthcheck, "command: %s", err.Error())
		}
		test = append([]string{string(check.Mode)}, command...)
	case docker.HealthcheckModeInherit, docker.HealthcheckModeCmdShell:
		test = []string{string(docker.HealthcheckModeCmdShell), check.Command}
	case docker.HealthcheckModeNone:
		test = []string{string(docker.HealthcheckModeNone)}
	default:
		return invalidBlock(specmodel.BlockContainerHealthcheck, "mode %q is not one of CMD, CMD-SHELL, NONE",
			check.Mode)
	}

	containerSpec.Healthcheck = &container.HealthConfig{
		Test:          test,
		Interval:      time.Duration(check.Interval),
		Timeout:       time.Duration(check.Timeout),
		StartPeriod:   time.Duration(check.StartPeriod),
		StartInterval: time.Duration(check.StartInterval),
		Retries:       check.Retries,
	}
	return nil
}

// buildResources truncates as the resource settings screen does.
func (s *service) buildResources(_ context.Context, state *buildState) error {
	resources := state.req.Doc.Deployment.Resources
	task := &state.req.Spec.TaskTemplate
	if task.Resources == nil {
		task.Resources = &swarm.ResourceRequirements{}
	}
	if reservations := resources.Reservations; reservations != nil {
		task.Resources.Reservations = &swarm.Resources{
			NanoCPUs:    docker.TruncateCPUsAsNano(reservations.CPUs, docker.MinCPUFraction),
			MemoryBytes: reservations.Memory.Truncate(unit.MB).Bytes(),
		}
	}
	if limits := resources.Limits; limits != nil {
		task.Resources.Limits = &swarm.Limit{
			NanoCPUs:    docker.TruncateCPUsAsNano(limits.CPUs, docker.MinCPUFraction),
			MemoryBytes: limits.Memory.Truncate(unit.MB).Bytes(),
			Pids:        limits.Pids,
		}
	}
	return nil
}

// buildStorage hands the mounts to volumeservice, which builds them the way the
// storage settings screen does.
//
// A mount's source here is a cluster-volume setting id. An export writes docker's
// source there instead - the volume name, or a host path once a bind volume was
// rewritten - which is why storage is left out of the export round-trip test.
// TODO: app templates phase 2 - map one onto the other before merging mounts.
// See docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
func (s *service) buildStorage(ctx context.Context, state *buildState) error {
	mounts := state.req.Doc.Deployment.Storage.Mounts
	requests := make([]*volumeservice.AppMountReq, 0, len(mounts))
	for _, target := range slices.Sorted(maps.Keys(mounts)) {
		m := mounts[target]
		// Options are always set: volumeservice applies the app's own subpath only
		// to a volume mount that carries them, and a mount without would share the
		// volume's root with every other app on it.
		options := &volumeservice.AppMountVolumeOptions{}
		if m.VolumeOptions != nil {
			options.Subpath = m.VolumeOptions.Subpath
		}
		requests = append(requests, &volumeservice.AppMountReq{
			Type:          m.Type,
			Source:        m.Source,
			Target:        target,
			ReadOnly:      m.ReadOnly,
			VolumeOptions: options,
		})
	}

	built, err := s.volumeService.BuildAppMounts(ctx, state.db, &volumeservice.BuildAppMountsReq{
		App: state.req.App,
		New: requests,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	state.req.Spec.TaskTemplate.ContainerSpec.Mounts = built.Mounts
	return nil
}
