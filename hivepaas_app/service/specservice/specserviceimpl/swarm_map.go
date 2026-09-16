package specserviceimpl

import (
	"strconv"
	"strings"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/services/docker"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

// mapSwarmService turns a live Swarm service into the declarative part of it.
//
// netNames maps Docker network id to name; an id with no entry is kept as-is,
// so an unresolved attachment is reported by import rather than lost here.
func mapSwarmService(
	svc *swarm.Service,
	netNames map[string]string,
) (*specmodel.Deployment, error) {
	if svc == nil {
		return nil, nil
	}
	spec := &svc.Spec
	task := &spec.TaskTemplate

	out := &specmodel.Deployment{
		Resources: mapResources(task),
		Networks:  mapNetworks(task, spec.EndpointSpec, netNames),
		Service:   mapService(spec, task),
	}

	if cs := task.ContainerSpec; cs != nil {
		out.Container = mapContainer(cs, task, spec.Labels)

		storage, err := mapStorage(cs)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		out.Storage = storage
	}
	return out, nil
}

// stripImageDigest removes the @sha256:… suffix docker stack deploy resolves an
// image to. It pins a digest that may not exist in the target registry and
// defeats the meaning of a moving tag. Services HivePaaS creates rarely carry
// one, since HivePaaS controls options.QueryRegistry - but it can be enabled.
func stripImageDigest(image string) string {
	if at := strings.LastIndex(image, "@"); at > 0 {
		return image[:at]
	}
	return image
}

func mapContainer(
	cs *swarm.ContainerSpec,
	task *swarm.TaskSpec,
	serviceLabels map[string]string,
) *specmodel.Container {
	out := &specmodel.Container{
		ServiceLabels:   filterUserLabels(serviceLabels),
		ContainerLabels: filterUserLabels(cs.Labels),
		Image:           stripImageDigest(cs.Image),
		Hostname:        cs.Hostname,
		User:            cs.User,
		Groups:          cs.Groups,
		StopSignal:      cs.StopSignal,
		TTY:             cs.TTY,
		Init:            cs.Init,
		OpenStdin:       cs.OpenStdin,
		ReadOnly:        cs.ReadOnly,
		Privileges:      mapPrivileges(cs.Privileges),
		Healthcheck:     mapHealthcheck(cs.Healthcheck),
		RestartPolicy:   mapRestartPolicy(task.RestartPolicy),
		LogDriver:       mapLogDriver(task.LogDriver),
	}
	if cs.StopGracePeriod != nil {
		out.StopGracePeriod = new(timeutil.Duration(*cs.StopGracePeriod))
	}
	return out
}

func mapPrivileges(privileges *swarm.Privileges) *specmodel.Privileges {
	if privileges == nil {
		return nil
	}
	out := &specmodel.Privileges{NoNewPrivileges: privileges.NoNewPrivileges}
	if ctx := privileges.SELinuxContext; ctx != nil {
		out.SELinuxContext = &specmodel.SELinuxContext{
			Disable: ctx.Disable, User: ctx.User, Role: ctx.Role, Type: ctx.Type, Level: ctx.Level,
		}
	}
	if seccomp := privileges.Seccomp; seccomp != nil {
		out.Seccomp = &specmodel.SeccompOpts{
			Mode: seccomp.Mode, Profile: string(seccomp.Profile),
		}
	}
	if appArmor := privileges.AppArmor; appArmor != nil {
		out.AppArmor = &specmodel.AppArmorOpts{Mode: appArmor.Mode}
	}
	return out
}

func mapHealthcheck(config *container.HealthConfig) *specmodel.Healthcheck {
	if config == nil {
		return nil
	}
	cmd := config.Test
	var mode docker.HealthcheckMode
	if len(cmd) > 0 {
		mode = docker.HealthcheckMode(cmd[0])
		cmd = cmd[1:]
	}
	return &specmodel.Healthcheck{
		Enabled:       mode != "NONE",
		Mode:          mode,
		Command:       strings.Join(cmd, " "),
		Interval:      timeutil.Duration(config.Interval),
		Timeout:       timeutil.Duration(config.Timeout),
		StartPeriod:   timeutil.Duration(config.StartPeriod),
		StartInterval: timeutil.Duration(config.StartInterval),
		Retries:       config.Retries,
	}
}

func mapRestartPolicy(policy *swarm.RestartPolicy) *specmodel.RestartPolicy {
	if policy == nil {
		return nil
	}
	out := &specmodel.RestartPolicy{
		Condition:   policy.Condition,
		MaxAttempts: policy.MaxAttempts,
	}
	if policy.Delay != nil {
		out.Delay = new(timeutil.Duration(*policy.Delay))
	}
	if policy.Window != nil {
		out.Window = new(timeutil.Duration(*policy.Window))
	}
	return out
}

func mapLogDriver(logDriver *swarm.Driver) *specmodel.LogDriver {
	if logDriver == nil {
		return nil
	}
	return &specmodel.LogDriver{Name: logDriver.Name, Options: logDriver.Options}
}

func mapResources(task *swarm.TaskSpec) *specmodel.Resources {
	out := &specmodel.Resources{
		Reservations: mapReservations(task.Resources),
		Limits:       mapLimits(task.Resources),
		Memory:       mapMemory(task),
	}
	if task.ContainerSpec != nil {
		out.Capabilities = mapCapabilities(task.ContainerSpec)
	}
	if out.Reservations == nil && out.Limits == nil && out.Memory == nil && out.Capabilities == nil {
		return nil
	}
	return out
}

func mapReservations(res *swarm.ResourceRequirements) *specmodel.ResourceReservations {
	if res == nil || res.Reservations == nil {
		return nil
	}
	out := &specmodel.ResourceReservations{
		CPUs:   float64(res.Reservations.NanoCPUs) / docker.UnitCPUNano,
		Memory: unit.DataSize(res.Reservations.MemoryBytes),
	}
	for _, r := range res.Reservations.GenericResources {
		switch {
		case r.NamedResourceSpec != nil:
			out.GenericResources = append(out.GenericResources, &specmodel.GenericResource{
				Kind: r.NamedResourceSpec.Kind, Value: r.NamedResourceSpec.Value,
			})
		case r.DiscreteResourceSpec != nil:
			out.GenericResources = append(out.GenericResources, &specmodel.GenericResource{
				Kind:  r.DiscreteResourceSpec.Kind,
				Value: strconv.FormatInt(r.DiscreteResourceSpec.Value, 10),
			})
		}
	}
	return out
}

func mapLimits(res *swarm.ResourceRequirements) *specmodel.ResourceLimits {
	if res == nil || res.Limits == nil {
		return nil
	}
	return &specmodel.ResourceLimits{
		CPUs:   float64(res.Limits.NanoCPUs) / docker.UnitCPUNano,
		Memory: unit.DataSize(res.Limits.MemoryBytes),
		Pids:   res.Limits.Pids,
	}
}

func mapMemory(task *swarm.TaskSpec) *specmodel.Memory {
	out := &specmodel.Memory{}
	if task.Resources != nil && task.Resources.SwapBytes != nil {
		out.Swap = new(unit.DataSize(*task.Resources.SwapBytes))
		out.Swappiness = task.Resources.MemorySwappiness
	}
	if shm := dockerhelper.GetShmMount(task); shm != nil && shm.TmpfsOptions != nil {
		out.ShmSize = new(unit.DataSize(shm.TmpfsOptions.SizeBytes))
	}
	if out.Swap == nil && out.Swappiness == nil && out.ShmSize == nil {
		return nil
	}
	return out
}

func mapCapabilities(cs *swarm.ContainerSpec) *specmodel.Capabilities {
	out := &specmodel.Capabilities{
		Ulimits: gofn.MapSlice(cs.Ulimits, func(u *container.Ulimit) *specmodel.Ulimit {
			return &specmodel.Ulimit{Name: u.Name, Hard: u.Hard, Soft: u.Soft}
		}),
		CapabilityAdd:  cs.CapabilityAdd,
		CapabilityDrop: cs.CapabilityDrop,
		EnableGPU:      gofn.Contain(cs.CapabilityAdd, "[gpu]"),
		OomScoreAdj:    cs.OomScoreAdj,
		Sysctls:        cs.Sysctls,
	}
	if len(out.Ulimits) == 0 && len(out.CapabilityAdd) == 0 && len(out.CapabilityDrop) == 0 &&
		!out.EnableGPU && out.OomScoreAdj == 0 && len(out.Sysctls) == 0 {
		return nil
	}
	return out
}

func mapStorage(cs *swarm.ContainerSpec) (*specmodel.Storage, error) {
	if len(cs.Mounts) == 0 {
		return nil, nil
	}
	mounts := make(map[string]specmodel.Mount, len(cs.Mounts))
	for i := range cs.Mounts {
		m := &cs.Mounts[i]
		if _, exists := mounts[m.Target]; exists {
			return nil, hperrors.Wrap(hperrors.ErrSpecMountTargetDuplicated).
				WithParam("Target", m.Target)
		}
		mounts[m.Target] = mapMount(m)
	}
	return &specmodel.Storage{Mounts: mounts}, nil
}

func mapMount(m *mount.Mount) specmodel.Mount {
	out := specmodel.Mount{
		Type:        m.Type,
		Source:      m.Source,
		ReadOnly:    m.ReadOnly,
		Consistency: m.Consistency,
	}
	if opts := m.BindOptions; opts != nil {
		out.BindOptions = &specmodel.BindOptions{
			Propagation:            opts.Propagation,
			NonRecursive:           opts.NonRecursive,
			CreateMountpoint:       opts.CreateMountpoint,
			ReadOnlyNonRecursive:   opts.ReadOnlyNonRecursive,
			ReadOnlyForceRecursive: opts.ReadOnlyForceRecursive,
		}
	}
	if opts := m.VolumeOptions; opts != nil {
		out.VolumeOptions = mapVolumeOptions(opts)
	}
	if opts := m.ClusterOptions; opts != nil {
		out.ClusterOptions = &specmodel.VolumeOptions{}
	}
	if opts := m.TmpfsOptions; opts != nil {
		out.TmpfsOptions = &specmodel.TmpfsOptions{
			Size:    unit.DataSize(opts.SizeBytes),
			Mode:    fileutil.FileMode(opts.Mode),
			Options: opts.Options,
		}
	}
	return out
}

func mapVolumeOptions(opts *mount.VolumeOptions) *specmodel.VolumeOptions {
	out := &specmodel.VolumeOptions{
		Subpath: opts.Subpath,
		NoCopy:  opts.NoCopy,
		Labels:  opts.Labels,
	}
	if driver := opts.DriverConfig; driver != nil {
		out.DriverConfig = &specmodel.VolumeDriver{Name: driver.Name, Options: driver.Options}
	}
	return out
}

func mapNetworks(
	task *swarm.TaskSpec,
	endpointSpec *swarm.EndpointSpec,
	netNames map[string]string,
) *specmodel.Networks {
	out := &specmodel.Networks{EndpointSpec: mapEndpointSpec(endpointSpec)}

	for _, attachment := range task.Networks {
		name := attachment.Target
		if resolved := netNames[attachment.Target]; resolved != "" {
			name = resolved
		}
		out.Attachments = append(out.Attachments, &specmodel.NetworkAttachment{
			Name: name, Aliases: attachment.Aliases,
		})
	}

	if cs := task.ContainerSpec; cs != nil {
		out.HostsFileEntries = mapHostsFileEntries(cs.Hosts)
		out.DNSConfig = mapDNSConfig(cs.DNSConfig)
	}

	if len(out.Attachments) == 0 && len(out.HostsFileEntries) == 0 &&
		out.DNSConfig == nil && out.EndpointSpec == nil {
		return nil
	}
	return out
}

func mapHostsFileEntries(hosts []string) []*specmodel.HostsFileEntry {
	out := make([]*specmodel.HostsFileEntry, 0, len(hosts))
	for _, host := range hosts {
		parts := gofn.StringSplit(host, " ", "\"")
		if len(parts) == 0 {
			continue
		}
		out = append(out, &specmodel.HostsFileEntry{Address: parts[0], Hostnames: parts[1:]})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func mapDNSConfig(config *swarm.DNSConfig) *specmodel.DNSConfig {
	if config == nil {
		return nil
	}
	nameservers := make([]string, 0, len(config.Nameservers))
	for i := range config.Nameservers {
		nameservers = append(nameservers, config.Nameservers[i].String())
	}
	return &specmodel.DNSConfig{
		Nameservers: nameservers, Search: config.Search, Options: config.Options,
	}
}

func mapEndpointSpec(endpointSpec *swarm.EndpointSpec) *specmodel.EndpointSpec {
	if endpointSpec == nil {
		return nil
	}
	out := &specmodel.EndpointSpec{Mode: endpointSpec.Mode}
	for _, port := range endpointSpec.Ports {
		out.Ports = append(out.Ports, &specmodel.PortConfig{
			Target:      port.TargetPort,
			Published:   port.PublishedPort,
			Protocol:    port.Protocol,
			PublishMode: port.PublishMode,
		})
	}
	return out
}

func mapService(spec *swarm.ServiceSpec, task *swarm.TaskSpec) *specmodel.Service {
	out := &specmodel.Service{ModeSpec: mapServiceMode(spec)}
	if task.Placement == nil {
		return out
	}
	// Platforms is detected by Docker at runtime and has no place in a
	// declarative document, so it is simply not mapped.
	constraints := filterUserConstraints(
		task.Placement.Constraints, spec.Labels[labelAppPlacementConstraints])
	if len(constraints) > 0 || len(task.Placement.Preferences) > 0 {
		out.Placement = &specmodel.Placement{
			Constraints: constraints,
			Preferences: mapPlacementPreferences(task.Placement.Preferences),
		}
	}
	return out
}

func mapServiceMode(spec *swarm.ServiceSpec) *specmodel.ServiceModeSpec {
	out := &specmodel.ServiceModeSpec{}
	switch {
	case spec.Mode.Replicated != nil:
		out.Mode = docker.ServiceModeReplicated
		out.ServiceReplicas = spec.Mode.Replicated.Replicas
	case spec.Mode.ReplicatedJob != nil:
		out.Mode = docker.ServiceModeReplicatedJob
		out.JobMaxConcurrent = spec.Mode.ReplicatedJob.MaxConcurrent
		out.JobTotalCompletions = spec.Mode.ReplicatedJob.TotalCompletions
	case spec.Mode.Global != nil:
		out.Mode = docker.ServiceModeGlobal
	case spec.Mode.GlobalJob != nil:
		out.Mode = docker.ServiceModeGlobalJob
	default:
		return nil
	}
	return out
}

func mapPlacementPreferences(prefs []swarm.PlacementPreference) []*specmodel.PlacementPreference {
	out := make([]*specmodel.PlacementPreference, 0, len(prefs))
	for _, pref := range prefs {
		if pref.Spread == nil {
			continue
		}
		out = append(out, &specmodel.PlacementPreference{
			Name: "spread", Value: pref.Spread.SpreadDescriptor,
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
