package loggingserviceimpl

import (
	"sort"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
	"github.com/hivepaas/hivepaas/services/logging"
)

// LabelManagedBy marks every service this package creates, so that finding them
// later does not depend on their names staying the same.
const LabelManagedBy = "hivepaas.logging.managed"

// logDriverLocal keeps the logging stack's own output away from the collector.
//
// The collector globs *-json.log; `local` writes a different format under
// local-logs/, so the stack's containers simply do not match. Excluding them by
// path is not possible - the path is a container id, never a service name - and
// collecting them feeds the collector its own output. `docker service logs`
// still reads them.
const logDriverLocal = "local"

type swarmSpecOpts struct {
	Name string
	// Global runs one task per node. The collector needs this; nothing else
	// HivePaaS creates uses it. See
	// docs/superpowers/notes/2026-09-12-global-mode-audit.md for what that was
	// checked against.
	Global bool
	// Constraint pins the service where its data is, in swarm's own spelling.
	// For the backend it is derived from the data volume, so the two can never
	// name different nodes.
	Constraint string
	// Resources caps what the task may take. Zero in a field is no cap.
	Resources logging.Resources
	Networks  []string
	// OomScoreAdj is the kernel OOM priority. It is applied only together with a
	// memory limit - see base.OomScoreAdjSystemAddon.
	OomScoreAdj int64
}

// toSwarmServiceSpec turns a description of a container into a swarm service.
//
// This is the only place in the logging subsystem that knows what swarm is.
func toSwarmServiceSpec(rt *logging.RuntimeSpec, opts swarmSpecOpts) (*swarm.ServiceSpec, error) {
	if rt.Image == "" {
		return nil, hperrors.Wrap(hperrors.ErrLoggingDeployFailed).
			WithExtraDetail("the runtime spec has no image")
	}

	container := &swarm.ContainerSpec{
		Image: rt.Image,
		Args:  rt.Args,
		Env:   envSlice(rt.Env),
	}
	for _, m := range rt.Mounts {
		container.Mounts = append(container.Mounts, toSwarmMount(m))
	}

	spec := &swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name:   opts.Name,
			Labels: map[string]string{LabelManagedBy: "true"},
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: container,
			RestartPolicy: &swarm.RestartPolicy{Condition: swarm.RestartPolicyConditionAny},
			LogDriver:     &swarm.Driver{Name: logDriverLocal},
		},
	}

	// Exactly one mode is set. The daemon rejects a spec carrying both, and a
	// Replicated left filled in "just in case" is how that happens.
	if opts.Global {
		spec.Mode.Global = &swarm.GlobalService{}
	} else {
		replicas := uint64(1)
		spec.Mode.Replicated = &swarm.ReplicatedService{Replicas: &replicas}
	}

	if opts.Constraint != "" {
		spec.TaskTemplate.Placement = &swarm.Placement{
			Constraints: []string{opts.Constraint},
		}
	}

	// Limits only. A reservation would make swarm refuse to schedule the task
	// at all on a busy node, and a logging backend that will not start is worse
	// than one sharing a node under pressure.
	if opts.Resources.CPULimit > 0 || opts.Resources.MemoryLimit > 0 {
		limits := &swarm.Limit{}
		if opts.Resources.CPULimit > 0 {
			limits.NanoCPUs = docker.TruncateCPUsAsNano(opts.Resources.CPULimit, docker.MinCPUFraction)
		}
		limits.MemoryBytes = opts.Resources.MemoryLimit
		spec.TaskTemplate.Resources = &swarm.ResourceRequirements{Limits: limits}
	}
	if opts.Resources.MemoryLimit > 0 {
		container.OomScoreAdj = opts.OomScoreAdj
	}

	for _, n := range opts.Networks {
		spec.TaskTemplate.Networks = append(spec.TaskTemplate.Networks,
			swarm.NetworkAttachmentConfig{Target: n})
	}

	return spec, nil
}

func toSwarmMount(m logging.Mount) mount.Mount {
	if m.VolumeName != "" {
		return mount.Mount{
			Type:     mount.TypeVolume,
			Source:   m.VolumeName,
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
		}
	}
	return mount.Mount{
		Type:     mount.TypeBind,
		Source:   m.Source,
		Target:   m.Target,
		ReadOnly: m.ReadOnly,
	}
}

// envSlice renders the environment in a stable order, so that the same
// configuration always produces the same spec. An unstable one would make every
// Apply look like a change and redeploy the service for nothing.
func envSlice(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, k+"="+env[k])
	}
	return out
}
