package specserviceimpl

import (
	"context"
	"maps"
	"slices"
	"strconv"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/docker"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
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
	applyResources(state.req.Doc.Deployment.Resources, &state.req.Spec.TaskTemplate)
	return nil
}

// applyResources writes the resources block onto a task the way the resource
// settings screen writes it, and replaces what the task held: a part the block
// leaves out is cleared.
func applyResources(r *specmodel.Resources, task *swarm.TaskSpec) {
	if r == nil {
		r = &specmodel.Resources{}
	}
	if task.Resources == nil {
		task.Resources = &swarm.ResourceRequirements{}
	}
	task.Resources.Reservations = buildReservations(r.Reservations)
	task.Resources.Limits = buildLimits(r.Limits)
	applyMemory(r.Memory, task)
	buildCapabilities(r.Capabilities, task)
}

// buildReservations reads a generic resource the way the resource settings
// screen does: a whole number is a count of something discrete, anything else
// names one.
func buildReservations(r *specmodel.ResourceReservations) *swarm.Resources {
	if r == nil {
		return nil
	}
	out := &swarm.Resources{
		NanoCPUs:    docker.TruncateCPUsAsNano(r.CPUs, docker.MinCPUFraction),
		MemoryBytes: r.Memory.Truncate(unit.MB).Bytes(),
	}
	for _, generic := range r.GenericResources {
		if generic == nil {
			continue
		}
		res := swarm.GenericResource{}
		if count, err := strconv.ParseInt(generic.Value, 10, 64); err == nil {
			res.DiscreteResourceSpec = &swarm.DiscreteGenericResource{Kind: generic.Kind, Value: count}
		} else {
			res.NamedResourceSpec = &swarm.NamedGenericResource{Kind: generic.Kind, Value: generic.Value}
		}
		out.GenericResources = append(out.GenericResources, res)
	}
	return out
}

func buildLimits(l *specmodel.ResourceLimits) *swarm.Limit {
	if l == nil {
		return nil
	}
	return &swarm.Limit{
		NanoCPUs:    docker.TruncateCPUsAsNano(l.CPUs, docker.MinCPUFraction),
		MemoryBytes: l.Memory.Truncate(unit.MB).Bytes(),
		Pids:        l.Pids,
	}
}

// applyMemory writes swap, swappiness and the size of /dev/shm, which is a tmpfs
// mount Docker is handed rather than a resource: without a size the mount goes.
func applyMemory(m *specmodel.Memory, task *swarm.TaskSpec) {
	task.Resources.SwapBytes, task.Resources.MemorySwappiness = nil, nil
	var shm *unit.DataSize
	if m != nil {
		if m.Swap != nil {
			task.Resources.SwapBytes = new(m.Swap.Truncate(unit.MB).Bytes())
		}
		if m.Swappiness != nil {
			task.Resources.MemorySwappiness = new(*m.Swappiness)
		}
		shm = m.ShmSize
	}
	if shm != nil && *shm > 0 {
		dockerhelper.SetShmSize(task, shm.Truncate(unit.MB).Bytes())
		return
	}
	if cs := task.ContainerSpec; cs != nil {
		if current := dockerhelper.GetShmMount(task); current != nil {
			target := current.Target
			cs.Mounts = slices.DeleteFunc(cs.Mounts, func(m mount.Mount) bool {
				return m.Type == mount.TypeTmpfs && m.Target == target
			})
		}
	}
}

// buildCapabilities writes the privileged part of the resources block onto the
// container, the way the app's resource settings screen writes it.
//
// Who is allowed to ask for this is decided before anything is built - a
// document carrying capabilities is provisioned only by someone who may change
// them - so there is nothing left to refuse here.
//
// The GPU is requested the way docker asks for it, by a capability spelled
// "[gpu]", which is why it is not a name the capability grammar would accept.
func buildCapabilities(capabilities *specmodel.Capabilities, task *swarm.TaskSpec) {
	if capabilities == nil {
		capabilities = &specmodel.Capabilities{}
	}
	contSpec := task.ContainerSpec
	contSpec.Ulimits = make([]*container.Ulimit, 0, len(capabilities.Ulimits))
	for _, ulimit := range capabilities.Ulimits {
		contSpec.Ulimits = append(contSpec.Ulimits,
			&container.Ulimit{Name: ulimit.Name, Hard: ulimit.Hard, Soft: ulimit.Soft})
	}
	contSpec.CapabilityAdd = slices.Clone(capabilities.CapabilityAdd)
	contSpec.CapabilityDrop = slices.Clone(capabilities.CapabilityDrop)
	if capabilities.EnableGPU {
		contSpec.CapabilityAdd = append(contSpec.CapabilityAdd, docker.CapabilityGPU)
	}
	contSpec.OomScoreAdj = capabilities.OomScoreAdj
	contSpec.Sysctls = maps.Clone(capabilities.Sysctls)
}

// buildNetworks publishes the ports the document asks for, the way the app's
// network settings screen publishes them.
//
// This is how an app answers something that is not HTTP: the reverse proxy
// serves the web addresses, and a VPN or a DNS server needs a port on the nodes
// themselves. Whether the port is free is decided before anything is built -
// docker would refuse it while creating the service, too late to say which port
// somebody asked for.
func (s *service) buildNetworks(_ context.Context, state *buildState) error {
	endpointSpec := state.req.Doc.Deployment.Networks.EndpointSpec
	if endpointSpec == nil {
		return nil
	}
	spec := state.req.Spec
	if spec.EndpointSpec == nil {
		spec.EndpointSpec = &swarm.EndpointSpec{}
	}
	spec.EndpointSpec.Mode = endpointSpec.Mode
	spec.EndpointSpec.Ports = make([]swarm.PortConfig, 0, len(endpointSpec.Ports))
	for _, port := range endpointSpec.Ports {
		spec.EndpointSpec.Ports = append(spec.EndpointSpec.Ports, swarm.PortConfig{
			TargetPort:    port.Target,
			PublishedPort: port.Published,
			Protocol:      port.Protocol,
			PublishMode:   port.PublishMode,
		})
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
			options.NoCopy = m.VolumeOptions.NoCopy
		}
		req := &volumeservice.AppMountReq{
			Type:          m.Type,
			Source:        m.Source,
			Target:        target,
			ReadOnly:      m.ReadOnly,
			VolumeOptions: options,
		}
		if err := s.applyMountSourceApp(ctx, state, m.SourceApp, req); err != nil {
			return hperrors.Wrap(err)
		}
		requests = append(requests, req)
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

// applyMountSourceApp points a mount at the directory of another app.
//
// The app is named by key, because that is what a template's `type: app`
// parameter carries and what a person reading the spec can recognize. Whether
// the caller may have that app's data was settled before provisioning started -
// this only has to find the app, and refuse a name that is not an app of this
// environment rather than quietly building the caller's own directory instead.
func (s *service) applyMountSourceApp(
	ctx context.Context,
	state *buildState,
	src *specmodel.MountSourceApp,
	out *volumeservice.AppMountReq,
) error {
	app := state.req.App
	if src == nil || src.App == "" || src.App == app.Key {
		return nil
	}

	apps, _, err := s.appRepo.List(ctx, state.db, app.ProjectID, nil,
		bunex.SelectWhere("app.project_env_id = ?", app.ProjectEnvID),
		bunex.SelectWhere("app.key = ?", src.App),
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if len(apps) == 0 {
		return hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail(
			"this environment has no app called %q for %s to mount", src.App, out.Target)
	}

	owner := apps[0]
	// The prefix is built from the project and the environment, which the row
	// alone does not carry; both are this app's own, since the app was found in
	// its environment.
	owner.Project, owner.ProjectEnv = app.Project, app.ProjectEnv
	out.OwnerApp = owner
	out.ReadOnly = !src.Write
	return nil
}

// buildInit decides whether docker puts an init process - tini - in front of
// the image's own command.
//
// An app is created with one, which reaps the orphans a process left behind
// and passes signals on. Some images bring their own supervisor that has to be
// process 1 itself: an s6-overlay image, which most of the linuxserver.io
// catalog and Firefly III are built on, refuses to start behind tini with
// "can only run as pid 1".
func (s *service) buildInit(_ context.Context, state *buildState) error {
	state.req.Spec.TaskTemplate.ContainerSpec.Init = state.req.Doc.Deployment.Container.Init
	return nil
}
