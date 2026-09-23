# Spec Import: Deployment Builders Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `specservice.BuildApp` can build every deployment block export writes, in an import
mode that checks a document with `CheckImportable` instead of the template caps, and whose
builders replace what a service spec holds rather than add to it.

**Architecture:** Each block's builder becomes a function of the block and the spec - `applyResources`,
`applyNetworks`, `applyContainer`, `applyService`, and a storage builder that also takes Docker
mounts - which writes the whole block and clears what the document leaves out. Templates keep
their gate and their block list; import gets `CheckImportable`, two blocks only it builds
(`deployment.container`, `deployment.service`), and `ImportBlocks`, which runs every deployment
builder whenever the document has a deployment. A round trip - build a document, read the spec
back with export's mapping - proves builder and export agree.

**Tech Stack:** Go 1.27, moby API types, testify.

**Spec:** [docs/superpowers/specs/2026-09-23-config-spec-import-design.md](../specs/2026-09-23-config-spec-import-design.md)
- §9 *Building apps*.

## Global Constraints

- Before calling a task done: `go build ./...`, `golangci-lint run ./...` over the whole repo, and
  `go test ./hivepaas_app/service/specservice/...` - the whole package, never a `-run` subset
  alone. The last task runs `go test ./...`.
- Templates build exactly what they built before: `CheckBuildable`, `PresentBlocks` and
  `BuildableBlocks` are unchanged, and every existing builder test passes untouched.
- A builder writes its whole block. A block, or a part of one, that the document leaves out is
  cleared - with two exceptions that keep what HivePaaS owns: network attachments are replaced
  only when the document names some, and a service mode is changed only when the document gives
  one.
- `container.image` is never written.
- Labels go through `dockerhelper.ApplyUserLabels`, placement keeps the constraints HivePaaS
  derived (those named in the `hivepaas.app.placementConstraints` label), and a missing log driver
  is `appservice.DefaultLogDriver()` - each what the settings screens do.
- In import mode a collection entry's `id` (`specmodel.CollectionEntryIDKey`) is removed before the
  entry is decoded, and the source may use any deployment method.
- End every commit message with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

---

## File Structure

| file | change |
|---|---|
| `specserviceimpl/build_deployment.go` | `applyResources`, `applyNetworks`, `applyHealthcheck`; storage builds Docker mounts, cluster mounts and consistency |
| `specserviceimpl/build_container.go` | new: `applyContainer` |
| `specserviceimpl/build_service.go` | new: `applyService` |
| `specserviceimpl/swarm_labels.go` | `managedConstraintSet`, `derivedConstraints` |
| `specserviceimpl/build.go` | import mode; the two new builders |
| `specserviceimpl/build_settings.go` | import mode: entry ids removed; any source method |
| `specserviceimpl/build_blocks_test.go` | new: the block builders' tests and round trips |
| `specserviceimpl/build_import_test.go` | new: import mode through `BuildApp` |
| `specmodel/importable.go` | new: `CheckImportable` |
| `specmodel/importable_test.go` | new |
| `specmodel/buildable.go` | `BlockContainer`, `BlockDeploymentService`, `ImportOnlyBlocks`, `ImportBlocks` |
| `specservice/types.go` | `BuildAppReq.Import` |

Paths below are relative to `hivepaas_app/service/specservice/`. Commands run from
`/Users/tnt/go/src/github.com/hivepaas/hivepaas`.

---

### Task 1: Resources are built whole

**Files:** `specserviceimpl/build_deployment.go` (`buildResources`, `buildCapabilities`); test
`specserviceimpl/build_blocks_test.go` (new).

**Produces:** `applyResources(r *specmodel.Resources, task *swarm.TaskSpec)`.

- [ ] **Step 1: Failing test.** Create `specserviceimpl/build_blocks_test.go`:

```go
package specserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func blankTask() *swarm.TaskSpec {
	return &swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}}
}

// Every field export reads back out of a task's resources, built and read again.
func TestResourcesRoundTrip(t *testing.T) {
	swap, shm, swappiness := unit.DataSize(512<<20), unit.DataSize(64<<20), int64(10)
	want := &specmodel.Resources{
		Reservations: &specmodel.ResourceReservations{
			CPUs: 0.5, Memory: 256 << 20,
			GenericResources: []*specmodel.GenericResource{{Kind: "gpu", Value: "2"}, {Kind: "ssd", Value: "fast"}},
		},
		Limits: &specmodel.ResourceLimits{CPUs: 1, Memory: 512 << 20, Pids: 100},
		Memory: &specmodel.Memory{Swap: &swap, Swappiness: &swappiness, ShmSize: &shm},
		Capabilities: &specmodel.Capabilities{
			CapabilityAdd: []string{"NET_ADMIN"}, Sysctls: map[string]string{"vm.max_map_count": "262144"},
			Ulimits: []*specmodel.Ulimit{{Name: "nofile", Soft: 1024, Hard: 2048}}, OomScoreAdj: -100,
		},
	}
	task := blankTask()

	applyResources(want, task)

	assert.Equal(t, want, mapResources(task))
}

// A block the document leaves out clears what the task held, so building an
// exported document over a running service converges on the document.
func TestResourcesLeftOutAreCleared(t *testing.T) {
	task := blankTask()
	applyResources(&specmodel.Resources{
		Limits:       &specmodel.ResourceLimits{CPUs: 1},
		Memory:       &specmodel.Memory{ShmSize: new(unit.DataSize(64 << 20))},
		Capabilities: &specmodel.Capabilities{CapabilityAdd: []string{"NET_ADMIN"}},
	}, task)

	applyResources(nil, task)

	assert.Nil(t, mapResources(task))
	assert.Empty(t, task.ContainerSpec.Mounts, "the shared-memory mount goes with memory.shmSize")
}
```

- [ ] **Step 2: Run** `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run TestResources -v` -
  expect `undefined: applyResources`.

- [ ] **Step 3: Implement.** In `build_deployment.go`, `buildResources` becomes a call to a
  function that writes the whole block (add `"strconv"` and
  `"github.com/hivepaas/hivepaas/services/docker/dockerhelper"` to the imports):

```go
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
```

  (`mount` is `"github.com/moby/moby/api/types/mount"`.) In `buildCapabilities`, replace
  `if capabilities == nil { return }` with `if capabilities == nil { capabilities = &specmodel.Capabilities{} }`,
  so a missing block clears the privileged part of the container too. Delete the old body of
  `buildResources`.

- [ ] **Step 4: Run** the whole spec package - expect PASS, the template builder tests included.
- [ ] **Step 5: Lint, commit** `feat(spec): build a resources block whole`.

---

### Task 2: Networks are built whole

**Files:** `specserviceimpl/build_deployment.go` (`buildNetworks`); test `build_blocks_test.go`.

**Produces:** `applyNetworks(n *specmodel.Networks, spec *swarm.ServiceSpec) error`.

- [ ] **Step 1: Failing tests.** Append to `build_blocks_test.go` (imports gain
  `"github.com/moby/moby/api/types/network"`):

```go
func blankSpec() *swarm.ServiceSpec {
	return &swarm.ServiceSpec{TaskTemplate: *blankTask()}
}

// Attachments name their network; Docker takes a name where it takes an id, which
// is how the network an app is created on is named too.
func TestNetworksRoundTrip(t *testing.T) {
	want := &specmodel.Networks{
		Attachments:      []*specmodel.NetworkAttachment{{Name: "shop_prod", Aliases: []string{"api"}}},
		HostsFileEntries: []*specmodel.HostsFileEntry{{Address: "10.0.0.5", Hostnames: []string{"db", "db.local"}}},
		DNSConfig:        &specmodel.DNSConfig{Nameservers: []string{"1.1.1.1"}, Search: []string{"local"}},
		EndpointSpec: &specmodel.EndpointSpec{Mode: swarm.ResolutionModeVIP, Ports: []*specmodel.PortConfig{
			{Target: 53, Published: 53, Protocol: network.UDP, PublishMode: swarm.PortConfigPublishModeHost},
		}},
	}
	spec := blankSpec()

	assert.NoError(t, applyNetworks(want, spec))

	assert.Equal(t, want, mapNetworks(&spec.TaskTemplate, spec.EndpointSpec, nil))
}

// The network an app was created on stays when the document names none: a
// template never does, and every app has one.
func TestNetworksKeepTheAttachmentsWhenTheDocumentNamesNone(t *testing.T) {
	spec := blankSpec()
	spec.TaskTemplate.Networks = []swarm.NetworkAttachmentConfig{{Target: "shop_prod", Aliases: []string{"web"}}}
	spec.TaskTemplate.ContainerSpec.Hosts = []string{"10.0.0.5 db"}

	assert.NoError(t, applyNetworks(nil, spec))

	assert.Len(t, spec.TaskTemplate.Networks, 1)
	assert.Empty(t, spec.TaskTemplate.ContainerSpec.Hosts)
	assert.Nil(t, spec.EndpointSpec)
}

func TestNetworksRefuseANameserverThatIsNotAnAddress(t *testing.T) {
	err := applyNetworks(&specmodel.Networks{DNSConfig: &specmodel.DNSConfig{Nameservers: []string{"dns.example"}}}, blankSpec())
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run** `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run TestNetworks -v` - expect
  `undefined: applyNetworks`.

- [ ] **Step 3: Implement.** `buildNetworks` becomes (imports gain `"net/netip"`, `"strings"`):

```go
// buildNetworks publishes the ports the document asks for, the way the app's
// network settings screen publishes them.
//
// This is how an app answers something that is not HTTP: the reverse proxy
// serves the web addresses, and a VPN or a DNS server needs a port on the nodes
// themselves. Whether the port is free is decided before anything is built -
// docker would refuse it while creating the service, too late to say which port
// somebody asked for.
func (s *service) buildNetworks(_ context.Context, state *buildState) error {
	return applyNetworks(state.req.Doc.Deployment.Networks, state.req.Spec)
}

// applyNetworks writes the networks block the way the network settings screen
// does, replacing what the spec held - except the attachments, which change only
// when the block names some: the network an app is created on is HivePaaS's to
// give, and a template never names it.
func applyNetworks(n *specmodel.Networks, spec *swarm.ServiceSpec) error {
	if n == nil {
		n = &specmodel.Networks{}
	}
	task := &spec.TaskTemplate
	if len(n.Attachments) > 0 {
		task.Networks = make([]swarm.NetworkAttachmentConfig, 0, len(n.Attachments))
		for _, attachment := range n.Attachments {
			if attachment == nil {
				continue
			}
			task.Networks = append(task.Networks, swarm.NetworkAttachmentConfig{
				Target: attachment.Name, Aliases: slices.Clone(attachment.Aliases),
			})
		}
	}

	cs := task.ContainerSpec
	cs.Hosts = nil
	for _, entry := range n.HostsFileEntries {
		if entry == nil {
			continue
		}
		cs.Hosts = append(cs.Hosts, strings.Join(append([]string{entry.Address}, entry.Hostnames...), " "))
	}

	cs.DNSConfig = nil
	if dns := n.DNSConfig; dns != nil {
		cs.DNSConfig = &swarm.DNSConfig{Search: slices.Clone(dns.Search), Options: slices.Clone(dns.Options)}
		for _, address := range dns.Nameservers {
			parsed, err := netip.ParseAddr(address)
			if err != nil {
				return invalidBlock(specmodel.BlockDeploymentNetworks, "dnsConfig: %q is not an address", address)
			}
			cs.DNSConfig.Nameservers = append(cs.DNSConfig.Nameservers, parsed)
		}
	}

	spec.EndpointSpec = nil
	if endpointSpec := n.EndpointSpec; endpointSpec != nil {
		spec.EndpointSpec = &swarm.EndpointSpec{Mode: endpointSpec.Mode}
		for _, port := range endpointSpec.Ports {
			if port == nil {
				continue
			}
			spec.EndpointSpec.Ports = append(spec.EndpointSpec.Ports, swarm.PortConfig{
				TargetPort: port.Target, PublishedPort: port.Published,
				Protocol: port.Protocol, PublishMode: port.PublishMode,
			})
		}
	}
	return nil
}
```

  `TestBuildAppPublishesThePortsTheDocumentAsksFor` compares `Ports` with a one-element slice,
  which the `append` form still produces.

- [ ] **Step 4: Run** the whole spec package - expect PASS.
- [ ] **Step 5: Lint, commit** `feat(spec): build a networks block whole`.

---

### Task 3: Storage takes Docker mounts, cluster mounts and consistency

**Files:** `specserviceimpl/build_deployment.go` (`buildStorage`; new `toDockerMount`); test
`build_blocks_test.go`.

**Produces:** `toDockerMount(target string, m specmodel.Mount) mount.Mount`.

- [ ] **Step 1: Failing test.** Append:

```go
// A Docker mount travels as Docker holds it, so building one is mapMount read
// backwards.
func TestDockerMountRoundTrip(t *testing.T) {
	for name, m := range map[string]mount.Mount{
		"a bind": {Type: mount.TypeBind, Source: "/etc/localtime", Target: "/etc/localtime", ReadOnly: true,
			BindOptions: &mount.BindOptions{Propagation: mount.PropagationRSlave, CreateMountpoint: true}},
		"a tmpfs": {Type: mount.TypeTmpfs, Target: "/cache",
			TmpfsOptions: &mount.TmpfsOptions{SizeBytes: 64 << 20, Mode: 0o1777}},
		"a volume mounted whole": {Type: mount.TypeVolume, Source: "shared", Target: "/shared",
			VolumeOptions: &mount.VolumeOptions{NoCopy: true, Labels: map[string]string{"a": "b"},
				DriverConfig: &mount.Driver{Name: "local", Options: map[string]string{"type": "nfs"}}}},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, m, toDockerMount(m.Target, mapMount(&m)))
		})
	}
}
```

  (imports gain `"github.com/moby/moby/api/types/mount"`.)

- [ ] **Step 2: Run** `-run TestDockerMountRoundTrip` - expect `undefined: toDockerMount`.

- [ ] **Step 3: Implement.** In `build_deployment.go`, `buildStorage` becomes (imports gain `"os"`):

```go
// buildStorage hands the mounts to volumeservice, which builds them the way the
// storage screen does. A managed mount's source is a cluster-volume setting id:
// a template names one, and import has resolved an exported volume to one before
// building. A Docker mount is kept as it is.
func (s *service) buildStorage(ctx context.Context, state *buildState) error {
	storage := state.req.Doc.Deployment.Storage
	if storage == nil {
		storage = &specmodel.Storage{}
	}
	requests := make([]*volumeservice.AppMountReq, 0, len(storage.Mounts))
	for _, target := range slices.Sorted(maps.Keys(storage.Mounts)) {
		m := storage.Mounts[target]
		req := &volumeservice.AppMountReq{
			Type:        m.Type,
			Source:      m.Source,
			Target:      target,
			ReadOnly:    m.ReadOnly,
			Consistency: m.Consistency,
		}
		// Options are always set: volumeservice applies the app's own subpath only
		// to a mount that carries them, and a mount without would share the
		// volume's root with every other app on it.
		opts, given := &volumeservice.AppMountVolumeOptions{}, m.VolumeOptions
		if m.Type == mount.TypeCluster {
			given = m.ClusterOptions
		}
		if given != nil {
			opts.Subpath, opts.NoCopy = given.Subpath, given.NoCopy
		}
		if m.Type == mount.TypeCluster {
			req.ClusterOptions = opts
		} else {
			req.VolumeOptions = opts
		}
		if err := s.applyMountSourceApp(ctx, state, m.SourceApp, req); err != nil {
			return hperrors.Wrap(err)
		}
		requests = append(requests, req)
	}

	kept := make([]mount.Mount, 0, len(storage.DockerMounts))
	for _, target := range slices.Sorted(maps.Keys(storage.DockerMounts)) {
		kept = append(kept, toDockerMount(target, storage.DockerMounts[target]))
	}

	built, err := s.volumeService.BuildAppMounts(ctx, state.db, &volumeservice.BuildAppMountsReq{
		App: state.req.App, Kept: kept, New: requests,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	state.req.Spec.TaskTemplate.ContainerSpec.Mounts = built.Mounts
	return nil
}

// toDockerMount is mapMount read backwards.
func toDockerMount(target string, m specmodel.Mount) mount.Mount {
	out := mount.Mount{
		Type: m.Type, Source: m.Source, Target: target, ReadOnly: m.ReadOnly, Consistency: m.Consistency,
	}
	if o := m.BindOptions; o != nil {
		out.BindOptions = &mount.BindOptions{
			Propagation: o.Propagation, NonRecursive: o.NonRecursive, CreateMountpoint: o.CreateMountpoint,
			ReadOnlyNonRecursive: o.ReadOnlyNonRecursive, ReadOnlyForceRecursive: o.ReadOnlyForceRecursive,
		}
	}
	if o := m.VolumeOptions; o != nil {
		out.VolumeOptions = &mount.VolumeOptions{Subpath: o.Subpath, NoCopy: o.NoCopy, Labels: o.Labels}
		if d := o.DriverConfig; d != nil {
			out.VolumeOptions.DriverConfig = &mount.Driver{Name: d.Name, Options: d.Options}
		}
	}
	if m.ClusterOptions != nil {
		out.ClusterOptions = &mount.ClusterOptions{}
	}
	if o := m.TmpfsOptions; o != nil {
		out.TmpfsOptions = &mount.TmpfsOptions{SizeBytes: o.Size.Bytes(), Mode: os.FileMode(o.Mode), Options: o.Options}
	}
	return out
}
```

  Delete the old comment's `TODO: app templates phase 2` paragraph with the old body - export now
  writes the builder's form. `TestBuildAppAlwaysGivesAVolumeMountOptions` and
  `TestBuildAppCarriesNoCopyToTheMount` keep passing: a volume mount still always carries options.

- [ ] **Step 4: Run** the whole spec package - expect PASS.
- [ ] **Step 5: Lint, commit** `feat(spec): build Docker mounts, cluster mounts and consistency`.

---

### Task 4: The container block

**Files:** `specserviceimpl/build_container.go` (new); `specserviceimpl/build_deployment.go`
(`buildHealthcheck` delegates); test `build_blocks_test.go`.

**Produces:** `applyContainer(c *specmodel.Container, spec *swarm.ServiceSpec) error` and
`applyHealthcheck(check *specmodel.Healthcheck, cs *swarm.ContainerSpec) error`.

- [ ] **Step 1: Failing tests.** Append (imports gain `"time"`,
  `"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"`,
  `"github.com/hivepaas/hivepaas/services/docker"`):

```go
func TestContainerRoundTrip(t *testing.T) {
	grace, delay := timeutil.Duration(20*time.Second), timeutil.Duration(5*time.Second)
	attempts := uint64(3)
	want := &specmodel.Container{
		ServiceLabels:   map[string]string{"team": "platform"},
		ContainerLabels: map[string]string{"tier": "db"},
		Image:           "placeholder",
		Hostname:        "db", User: "999", Groups: []string{"audio"}, StopSignal: "SIGINT",
		TTY: true, Init: new(true), OpenStdin: true, ReadOnly: true, StopGracePeriod: &grace,
		Privileges: &specmodel.Privileges{
			NoNewPrivileges: true,
			SELinuxContext:  &specmodel.SELinuxContext{Type: "container_t"},
			Seccomp:         &specmodel.SeccompOpts{Mode: swarm.SeccompModeUnconfined},
			AppArmor:        &specmodel.AppArmorOpts{Mode: swarm.AppArmorModeDisabled},
		},
		Healthcheck: &specmodel.Healthcheck{Enabled: true, Mode: docker.HealthcheckModeCmdShell,
			Command: "pg_isready", Interval: timeutil.Duration(10 * time.Second), Retries: 5},
		RestartPolicy: &specmodel.RestartPolicy{Condition: swarm.RestartPolicyConditionOnFailure,
			Delay: &delay, MaxAttempts: &attempts},
		LogDriver: &specmodel.LogDriver{Name: "json-file", Options: map[string]string{"max-size": "10m"}},
	}
	spec := blankSpec()
	spec.TaskTemplate.ContainerSpec.Image = "placeholder"

	assert.NoError(t, applyContainer(want, spec))

	assert.Equal(t, want, mapContainer(spec.TaskTemplate.ContainerSpec, &spec.TaskTemplate, spec.Labels))
}

// The image changes only through a deployment, and labels HivePaaS put on the
// service stay whatever the document says.
func TestContainerKeepsTheImageAndHivePaaSLabels(t *testing.T) {
	spec := blankSpec()
	spec.Labels = map[string]string{"hivepaas.app.info": "x", "team": "old"}
	spec.TaskTemplate.ContainerSpec.Image = "registry/app:running"

	assert.NoError(t, applyContainer(&specmodel.Container{Image: "registry/app:exported",
		ServiceLabels: map[string]string{"team": "new"}}, spec))

	assert.Equal(t, "registry/app:running", spec.TaskTemplate.ContainerSpec.Image)
	assert.Equal(t, map[string]string{"hivepaas.app.info": "x", "team": "new"}, spec.Labels)
}

// Left out, the block clears the container back to what HivePaaS creates an app
// with: HivePaaS's log driver, no healthcheck of its own.
func TestContainerLeftOutIsCleared(t *testing.T) {
	spec := blankSpec()
	assert.NoError(t, applyContainer(&specmodel.Container{Hostname: "x", User: "1",
		Healthcheck: &specmodel.Healthcheck{Enabled: true, Command: "true"}}, spec))

	assert.NoError(t, applyContainer(nil, spec))

	cs := spec.TaskTemplate.ContainerSpec
	assert.Empty(t, cs.Hostname)
	assert.Empty(t, cs.User)
	assert.Nil(t, cs.Healthcheck)
	assert.NotNil(t, spec.TaskTemplate.LogDriver)
}

// NONE turns the image's own healthcheck off, which is not the same as leaving
// the image's healthcheck on.
func TestHealthcheckNoneStaysOff(t *testing.T) {
	cs := &swarm.ContainerSpec{}
	assert.NoError(t, applyHealthcheck(&specmodel.Healthcheck{Mode: docker.HealthcheckModeNone}, cs))
	assert.Equal(t, []string{"NONE"}, cs.Healthcheck.Test)
}
```

- [ ] **Step 2: Run** `-run 'TestContainer|TestHealthcheck'` - expect `undefined: applyContainer`.

- [ ] **Step 3: Implement.** Move the body of `buildHealthcheck` into `applyHealthcheck`, which
  `buildHealthcheck` now calls with `state.req.Doc.Deployment.Container.Healthcheck`; a nil check
  sets `Healthcheck = nil`, and a disabled check sets `Healthcheck = nil` unless its mode is
  `NONE`, which becomes `Test: ["NONE"]`:

```go
func (s *service) buildHealthcheck(_ context.Context, state *buildState) error {
	return applyHealthcheck(state.req.Doc.Deployment.Container.Healthcheck,
		state.req.Spec.TaskTemplate.ContainerSpec)
}

// applyHealthcheck writes a healthcheck in the form docker runs it. CMD is an
// argv, so its command is split; CMD-SHELL is one string handed to the shell,
// so its command is not. A mode left empty means CMD-SHELL, which is what a
// template author writing a shell command expects. NONE turns the image's own
// healthcheck off, where no healthcheck at all leaves the image's.
//
// The container settings screen splits a CMD-SHELL command too, and docker then
// hands the shell only its first word: `sh -c pg_isready -U app` runs pg_isready
// with no arguments. That is a bug to fix there, not a behavior to copy here.
func applyHealthcheck(check *specmodel.Healthcheck, containerSpec *swarm.ContainerSpec) error {
	if check == nil || (!check.Enabled && check.Mode != docker.HealthcheckModeNone) {
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
```

  Create `specserviceimpl/build_container.go`:

```go
package specserviceimpl

import (
	"maps"
	"slices"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

// applyContainer writes the container block the way the container settings
// screen does, replacing what the spec held. Three things stay HivePaaS's: the
// image, which only a deployment changes; the labels HivePaaS put on the service
// and the container, which ApplyUserLabels keeps; and the log driver, which is
// HivePaaS's default when the block names none.
func applyContainer(c *specmodel.Container, spec *swarm.ServiceSpec) error {
	if c == nil {
		c = &specmodel.Container{}
	}
	task := &spec.TaskTemplate
	cs := task.ContainerSpec

	spec.Labels = dockerhelper.ApplyUserLabels(spec.Labels, c.ServiceLabels)
	cs.Labels = dockerhelper.ApplyUserLabels(cs.Labels, c.ContainerLabels)
	cs.Hostname, cs.User, cs.StopSignal = c.Hostname, c.User, c.StopSignal
	cs.Groups = slices.Clone(c.Groups)
	cs.TTY, cs.OpenStdin, cs.ReadOnly = c.TTY, c.OpenStdin, c.ReadOnly
	cs.Init = nil
	if c.Init != nil {
		cs.Init = new(*c.Init)
	}
	cs.StopGracePeriod = nil
	if c.StopGracePeriod != nil {
		cs.StopGracePeriod = new(time.Duration(*c.StopGracePeriod))
	}
	cs.Privileges = toDockerPrivileges(c.Privileges)
	if err := applyHealthcheck(c.Healthcheck, cs); err != nil {
		return err
	}
	task.RestartPolicy = toDockerRestartPolicy(c.RestartPolicy)
	task.LogDriver = appservice.DefaultLogDriver()
	if c.LogDriver != nil && c.LogDriver.Name != "" {
		task.LogDriver = &swarm.Driver{Name: c.LogDriver.Name, Options: maps.Clone(c.LogDriver.Options)}
	}
	return nil
}

func toDockerPrivileges(p *specmodel.Privileges) *swarm.Privileges {
	if p == nil {
		return nil
	}
	out := &swarm.Privileges{NoNewPrivileges: p.NoNewPrivileges}
	if ctx := p.SELinuxContext; ctx != nil {
		out.SELinuxContext = &swarm.SELinuxContext{
			Disable: ctx.Disable, User: ctx.User, Role: ctx.Role, Type: ctx.Type, Level: ctx.Level,
		}
	}
	if seccomp := p.Seccomp; seccomp != nil {
		out.Seccomp = &swarm.SeccompOpts{Mode: seccomp.Mode}
		if seccomp.Profile != "" {
			out.Seccomp.Profile = []byte(seccomp.Profile)
		}
	}
	if appArmor := p.AppArmor; appArmor != nil {
		out.AppArmor = &swarm.AppArmorOpts{Mode: appArmor.Mode}
	}
	return out
}

func toDockerRestartPolicy(p *specmodel.RestartPolicy) *swarm.RestartPolicy {
	if p == nil {
		return nil
	}
	out := &swarm.RestartPolicy{Condition: p.Condition}
	if p.MaxAttempts != nil {
		out.MaxAttempts = new(*p.MaxAttempts)
	}
	if p.Delay != nil {
		out.Delay = new(time.Duration(*p.Delay))
	}
	if p.Window != nil {
		out.Window = new(time.Duration(*p.Window))
	}
	return out
}
```

- [ ] **Step 4: Run** the whole spec package - expect PASS.
- [ ] **Step 5: Lint, commit** `feat(spec): build the container block`.

---

### Task 5: The service block

**Files:** `specserviceimpl/build_service.go` (new); `specserviceimpl/swarm_labels.go`; test
`build_blocks_test.go`.

**Produces:** `applyService(s *specmodel.Service, spec *swarm.ServiceSpec)`.

- [ ] **Step 1: Failing tests.** Append:

```go
func TestServiceRoundTrip(t *testing.T) {
	replicas := uint64(3)
	want := &specmodel.Service{
		ModeSpec: &specmodel.ServiceModeSpec{Mode: docker.ServiceModeReplicated, ServiceReplicas: &replicas},
		Placement: &specmodel.Placement{
			Constraints: []string{"node.labels.zone==eu"},
			Preferences: []*specmodel.PlacementPreference{{Name: "spread", Value: "node.labels.rack"}},
		},
	}
	spec := blankSpec()

	applyService(want, spec)

	assert.Equal(t, want, mapService(spec, &spec.TaskTemplate))
}

// A constraint HivePaaS derived - the pin to the node a volume lives on - is
// regenerated at the next deployment, but until then it is what keeps the task
// beside its data, so building the block keeps it.
func TestServiceKeepsTheConstraintsHivePaaSDerived(t *testing.T) {
	spec := blankSpec()
	spec.Labels = map[string]string{labelAppPlacementConstraints: "node.id==abc"}
	spec.TaskTemplate.Placement = &swarm.Placement{Constraints: []string{"node.id == abc", "node.labels.zone==us"}}

	applyService(&specmodel.Service{Placement: &specmodel.Placement{Constraints: []string{"node.labels.zone==eu"}}}, spec)

	assert.Equal(t, []string{"node.id == abc", "node.labels.zone==eu"}, spec.TaskTemplate.Placement.Constraints)
}

// Without a mode the service keeps its own: swarm refuses to change the kind of
// service in place, and nothing here says what it should become.
func TestServiceWithoutAModeKeepsItsOwn(t *testing.T) {
	spec := blankSpec()
	spec.Mode = swarm.ServiceMode{Global: &swarm.GlobalService{}}

	applyService(nil, spec)

	assert.NotNil(t, spec.Mode.Global)
}
```

- [ ] **Step 2: Run** `-run TestService` - expect `undefined: applyService`.

- [ ] **Step 3: Implement.** In `swarm_labels.go`, factor the parsing out of
  `filterUserConstraints`:

```go
// managedConstraintSet is the placement constraints HivePaaS derived, as the
// label listing them says, each normalized.
func managedConstraintSet(managed string) []string {
	set := make([]string, 0, 4) //nolint:mnd
	for _, item := range strings.Split(managed, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			set = append(set, normalizeConstraint(trimmed))
		}
	}
	return set
}

// derivedConstraints is the constraints of a placement HivePaaS derived: the ones
// export leaves out, and building the block keeps.
func derivedConstraints(placement *swarm.Placement, managed string) []string {
	if placement == nil {
		return nil
	}
	set := managedConstraintSet(managed)
	var out []string
	for _, constraint := range placement.Constraints {
		if gofn.Contain(set, normalizeConstraint(constraint)) {
			out = append(out, constraint)
		}
	}
	return out
}
```

  and `filterUserConstraints` uses `managedSet := managedConstraintSet(managed)` in place of its
  loop (import `"github.com/moby/moby/api/types/swarm"`). Create `specserviceimpl/build_service.go`:

```go
package specserviceimpl

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/services/docker"
)

// applyService writes the service block the way the service settings screen
// does. The mode changes only when the block gives one; the placement is the
// block's constraints after the ones HivePaaS derived, which stay.
func applyService(svc *specmodel.Service, spec *swarm.ServiceSpec) {
	if svc == nil {
		svc = &specmodel.Service{}
	}
	applyServiceMode(svc.ModeSpec, spec)

	task := &spec.TaskTemplate
	constraints := derivedConstraints(task.Placement, spec.Labels[labelAppPlacementConstraints])
	var preferences []swarm.PlacementPreference
	if p := svc.Placement; p != nil {
		constraints = append(constraints, p.Constraints...)
		for _, pref := range p.Preferences {
			if pref != nil && pref.Name == "spread" {
				preferences = append(preferences, swarm.PlacementPreference{
					Spread: &swarm.SpreadOver{SpreadDescriptor: pref.Value},
				})
			}
		}
	}
	if task.Placement == nil {
		task.Placement = &swarm.Placement{}
	}
	task.Placement.Constraints, task.Placement.Preferences = constraints, preferences
}

func applyServiceMode(m *specmodel.ServiceModeSpec, spec *swarm.ServiceSpec) {
	if m == nil || m.Mode == "" {
		return
	}
	spec.Mode = swarm.ServiceMode{}
	switch m.Mode {
	case docker.ServiceModeReplicated:
		spec.Mode.Replicated = &swarm.ReplicatedService{Replicas: copyUint(m.ServiceReplicas)}
	case docker.ServiceModeReplicatedJob:
		spec.Mode.ReplicatedJob = &swarm.ReplicatedJob{
			MaxConcurrent: copyUint(m.JobMaxConcurrent), TotalCompletions: copyUint(m.JobTotalCompletions),
		}
	case docker.ServiceModeGlobal:
		spec.Mode.Global = &swarm.GlobalService{}
	case docker.ServiceModeGlobalJob:
		spec.Mode.GlobalJob = &swarm.GlobalJob{}
	}
}

func copyUint(v *uint64) *uint64 {
	if v == nil {
		return nil
	}
	return new(*v)
}
```

- [ ] **Step 4: Run** the whole spec package - expect PASS.
- [ ] **Step 5: Lint, commit** `feat(spec): build the service block`.

---

### Task 6: Import mode

**Files:** `specmodel/buildable.go`, `specmodel/importable.go` (new), `specmodel/importable_test.go`
(new), `specservice/types.go`, `specserviceimpl/build.go`, `specserviceimpl/build_settings.go`,
`specserviceimpl/build_deployment.go` (`buildSource`), `specserviceimpl/build_test.go`
(`TestBuilderRegistryCoversEveryBuildableBlock`), `specserviceimpl/build_import_test.go` (new).

**Produces:** `specmodel.BlockContainer`, `specmodel.BlockDeploymentService`,
`specmodel.ImportOnlyBlocks`, `specmodel.ImportBlocks(doc) []Block`,
`specmodel.CheckImportable(doc) error`, `specservice.BuildAppReq.Import bool`.

- [ ] **Step 1: Failing tests.** Create `specmodel/importable_test.go`:

```go
package specmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func TestCheckImportableAcceptsWhatATemplateCannotSay(t *testing.T) {
	doc := decodeDoc(t, `
deployment:
  container: {user: "999", readOnly: true}
  service: {modeSpec: {mode: global}}
  resources: {memory: {swap: 1gb}, reservations: {genericResources: [{kind: gpu, value: '1'}]}}
  networks: {dnsConfig: {nameservers: [1.1.1.1]}, attachments: [{name: shop_prod}]}
  storage:
    mounts: {/data: {type: cluster, source: vol-1, clusterOptions: {subpath: data}}}
    dockerMounts: {/cache: {type: tmpfs}}
`)
	assert.NoError(t, CheckImportable(doc))
}

func TestCheckImportableRefuses(t *testing.T) {
	cases := map[string]string{
		"a volume import has not resolved": "deployment:\n  storage:\n    mounts:\n" +
			"      /data: {type: volume, external: {type: cluster-volume, name: v}}\n",
		"a managed mount with no volume": "deployment:\n  storage:\n    mounts:\n      /data: {type: volume}\n",
		"a managed bind":                 "deployment:\n  storage:\n    mounts:\n      /data: {type: bind, source: v}\n",
		"a relative target":              "deployment:\n  storage:\n    dockerMounts:\n      data: {type: tmpfs}\n",
		"a target in both maps": "deployment:\n  storage:\n    mounts:\n      /d: {type: volume, source: v}\n" +
			"    dockerMounts:\n      /d: {type: tmpfs}\n",
		"a subpath leaving the directory": "deployment:\n  storage:\n    mounts:\n" +
			"      /d: {type: volume, source: v, volumeOptions: {subpath: ../x}}\n",
		"an unknown mode":  "deployment:\n  service: {modeSpec: {mode: sometimes}}\n",
		"a port too large": "deployment:\n  networks: {endpointSpec: {ports: [{target: 70000}]}}\n",
	}
	for name, doc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.ErrorIs(t, CheckImportable(decodeDoc(t, doc)), hperrors.ErrSpecBlockInvalid)
		})
	}
}

// Every deployment block is built for an export with a deployment - a block it
// left out is one the app has none of - and the source only when it is there.
func TestImportBlocks(t *testing.T) {
	assert.Equal(t, []Block{
		BlockDeploymentStorage, BlockContainer, BlockDeploymentResources, BlockDeploymentNetworks,
		BlockDeploymentService, BlockSettingsKind,
	}, ImportBlocks(decodeDoc(t, "deployment:\n  container: {}\nsettings:\n  kind: {category: webapp}\n")))
	assert.Equal(t, []Block{BlockSettingsRouting},
		ImportBlocks(decodeDoc(t, "settings:\n  routing: {port: 80}\n")), "an app never deployed")
}
```

  Create `specserviceimpl/build_import_test.go`:

```go
package specserviceimpl

import (
	"context"
	"slices"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

const importDocYAML = `
deployment:
  source:
    activeMethod: repo
    repoSource: {repoUrl: "https://github.com/acme/api", repoRef: main}
  container:
    serviceLabels: {team: platform}
    hostname: api
    user: "1000"
    readOnly: true
    healthcheck: {enabled: true, mode: CMD-SHELL, command: curl -f localhost, retries: 3}
    logDriver: {name: json-file, options: {max-size: 10m}}
  resources:
    limits: {cpus: 1, memory: 512mb}
    memory: {shmSize: 64mb}
  networks:
    attachments: [{name: shop_prod, aliases: [api]}]
    hostsFileEntries: [{address: 10.0.0.5, hostnames: [db]}]
  service:
    modeSpec: {mode: replicated, serviceReplicas: 2}
    placement: {constraints: ["node.labels.zone==eu"]}
  storage:
    mounts:
      /data: {type: volume, source: vol-1, volumeOptions: {subpath: data}}
    dockerMounts:
      /etc/localtime: {type: bind, source: /etc/localtime, readOnly: true}
settings:
  secrets:
    DB_PASSWORD: {id: 01JSECRET, key: DB_PASSWORD, value: hunter2}
`

// An exported document, built in import mode and read back with export's own
// mapping, comes back as it went in: builder and export agree about every block.
func TestBuildAppImportsAnExportedDocument(t *testing.T) {
	useDataKey(t)
	volumes := &fakeBuildVolumeService{}
	svc := &service{volumeService: volumes}
	req := buildReq(t, importDocYAML)
	req.Import = true

	resp, err := svc.BuildApp(context.Background(), nil, req)

	assert.NoError(t, err)
	out := mapSwarmService(&swarm.Service{Spec: *req.Spec}, nil)
	doc := req.Doc.Deployment
	doc.Container.Image = req.Spec.TaskTemplate.ContainerSpec.Image // never written
	assert.Equal(t, doc.Container, out.Container)
	assert.Equal(t, doc.Networks, out.Networks)
	assert.Equal(t, doc.Service, out.Service)
	assert.Equal(t, doc.Resources, out.Resources)

	if assert.Len(t, volumes.req.New, 1) {
		assert.Equal(t, "vol-1", volumes.req.New[0].Source)
		assert.Equal(t, "data", volumes.req.New[0].VolumeOptions.Subpath)
	}
	assert.Equal(t, []mount.Mount{{Type: mount.TypeBind, Source: "/etc/localtime", Target: "/etc/localtime",
		ReadOnly: true}}, volumes.req.Kept)

	secret := slices.IndexFunc(resp.Settings, func(s *entity.Setting) bool { return s.Type == base.SettingTypeSecret })
	assert.GreaterOrEqual(t, secret, 0, "the entry's id is taken out before it is decoded")
}

// A template still cannot say what only an export can.
func TestBuildAppKeepsTheTemplateGate(t *testing.T) {
	svc := &service{volumeService: &fakeBuildVolumeService{}}
	_, err := svc.BuildApp(context.Background(), nil, buildReq(t, "deployment:\n  container: {user: root}\n"))
	assert.Error(t, err)
}

// An export with no deployment - an app never deployed - builds no service block.
func TestBuildAppImportsAnAppNeverDeployed(t *testing.T) {
	useDataKey(t)
	svc := &service{volumeService: &fakeBuildVolumeService{}}
	req := buildReq(t, "settings:\n  routing: {port: 80}\n")
	req.Import = true

	_, err := svc.BuildApp(context.Background(), nil, req)

	assert.NoError(t, err)
	assert.Equal(t, "busybox:latest", req.Spec.TaskTemplate.ContainerSpec.Image)
}
```

- [ ] **Step 2: Run** the whole spec package - expect compile failures naming `CheckImportable`,
  `ImportBlocks`, `BlockContainer`, `req.Import`.

- [ ] **Step 3: Implement.**

  *`specmodel/buildable.go`* - after the block constants:

```go
// Blocks only import builds. A template cannot ask for them - CheckBuildable
// refuses every field they carry - so PresentBlocks never names them.
const (
	BlockContainer         Block = "deployment.container"
	BlockDeploymentService Block = "deployment.service"
)

var ImportOnlyBlocks = []Block{BlockContainer, BlockDeploymentService}

// buildOrder is the order blocks are built in. Storage replaces the mounts
// before resources sets the size of /dev/shm, which is one of them.
var buildOrder = []Block{
	BlockDeploymentSource, BlockDeploymentStorage, BlockContainerHealthcheck, BlockContainerInit,
	BlockContainer, BlockDeploymentResources, BlockDeploymentNetworks, BlockDeploymentService,
	BlockSettingsKind, BlockSettingsEnvVars, BlockSettingsSecrets, BlockSettingsConfigFiles,
	BlockSettingsRouting,
}

// ImportBlocks lists the blocks an exported document is built with, in build
// order. Every deployment block is built when the document has a deployment at
// all: export writes a block only when it holds something, so a block missing
// from a deployment is one the app has none of, and building it clears whatever
// the service holds. The container block covers its healthcheck and init. The
// source is a setting rather than part of the service, and is built only when
// present, as settings are - import never deletes a setting.
func ImportBlocks(doc *AppDoc) []Block {
	var blocks []Block
	if doc == nil {
		return blocks
	}
	if d := doc.Deployment; d != nil {
		if d.Source != nil {
			blocks = append(blocks, BlockDeploymentSource)
		}
		blocks = append(blocks, BlockDeploymentStorage, BlockContainer, BlockDeploymentResources,
			BlockDeploymentNetworks, BlockDeploymentService)
	}
	blocks = append(blocks, presentSettingsBlocks(doc)...)
	slices.SortFunc(blocks, func(a, b Block) int {
		return slices.Index(buildOrder, a) - slices.Index(buildOrder, b)
	})
	return blocks
}
```

  and move the two settings loops out of `PresentBlocks` into
  `func presentSettingsBlocks(doc *AppDoc) []Block`, which `PresentBlocks` appends before sorting.

  *`specmodel/importable.go`* (new):

```go
package specmodel

import (
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
)

var serviceModes = []docker.ServiceMode{
	docker.ServiceModeReplicated, docker.ServiceModeReplicatedJob,
	docker.ServiceModeGlobal, docker.ServiceModeGlobalJob,
}

// CheckImportable refuses an exported document that cannot be built as it
// stands. It holds none of a template's caps - the document describes what an
// installation already ran - only what building needs: every volume resolved to
// a setting, every mount at an absolute target of its own, a subpath that stays
// below its directory, a mode swarm knows, ports that are ports.
func CheckImportable(doc *AppDoc) error {
	if doc == nil || doc.Deployment == nil {
		return nil
	}
	d := doc.Deployment
	if err := checkImportedStorage(d.Storage); err != nil {
		return err
	}
	if s := d.Service; s != nil && s.ModeSpec != nil && s.ModeSpec.Mode != "" &&
		!slices.Contains(serviceModes, s.ModeSpec.Mode) {
		return invalid("deployment.service.modeSpec.mode: %q is not a mode swarm knows", s.ModeSpec.Mode)
	}
	if n := d.Networks; n != nil && n.EndpointSpec != nil {
		for i, port := range n.EndpointSpec.Ports {
			if port != nil && (port.Target < 1 || port.Target > maxPortNumber || port.Published > maxPortNumber) {
				return invalid("deployment.networks.endpointSpec.ports[%d]: not a port", i)
			}
		}
	}
	return nil
}

func checkImportedStorage(s *Storage) error {
	if s == nil {
		return nil
	}
	for _, target := range slices.Sorted(maps.Keys(s.Mounts)) {
		m, at := s.Mounts[target], "deployment.storage.mounts."+target
		switch {
		case !path.IsAbs(target):
			return invalid("%s: the target is not an absolute path", at)
		case m.External != nil:
			return invalid("%s: the volume %q has to be resolved before the app is built", at, m.External.Name)
		case m.Type != mount.TypeVolume && m.Type != mount.TypeCluster:
			return invalid("%s: a managed mount is a volume or a cluster volume, not %q", at, m.Type)
		case m.Source == "":
			return invalid("%s: no volume", at)
		}
		for _, opts := range []*VolumeOptions{m.VolumeOptions, m.ClusterOptions} {
			if opts != nil && !staysBelow(opts.Subpath) {
				return invalid("%s: the subpath %q leaves the app's directory", at, opts.Subpath)
			}
		}
		if _, twice := s.DockerMounts[target]; twice {
			return invalid("%s: the target is mounted twice", at)
		}
	}
	for _, target := range slices.Sorted(maps.Keys(s.DockerMounts)) {
		if !path.IsAbs(target) {
			return invalid("deployment.storage.dockerMounts.%s: the target is not an absolute path", target)
		}
	}
	return nil
}

// staysBelow reports whether a subpath names a directory below the one it is
// joined to.
func staysBelow(subpath string) bool {
	if subpath == "" {
		return true
	}
	cleaned := path.Clean(subpath)
	return !path.IsAbs(cleaned) && cleaned != ".." && !strings.HasPrefix(cleaned, "../")
}

func invalid(format string, args ...any) error {
	return hperrors.Wrap(hperrors.ErrSpecBlockInvalid).WithExtraDetail("%s", fmt.Sprintf(format, args...))
}
```

  *`specservice/types.go`* - `BuildAppReq` gains, after `TimeNow`:

```go
	// Import says the document is an export rather than a template: it is checked
	// with CheckImportable, every block export writes is built, and each block
	// replaces what Spec holds.
	Import bool
```

  *`specserviceimpl/build.go`* - `BuildApp` picks its gate and block list:

```go
	check, blocks := specmodel.CheckBuildable, specmodel.PresentBlocks
	if req.Import {
		check, blocks = specmodel.CheckImportable, specmodel.ImportBlocks
	}
	if err := check(req.Doc); err != nil {
		return nil, hperrors.Wrap(err)
	}

	builders := s.builders()
	state := &buildState{db: db, req: req}
	for _, block := range blocks(req.Doc) {
```

  and `builders()` gains:

```go
		specmodel.BlockContainer:            s.buildContainer,
		specmodel.BlockDeploymentService:    s.buildService,
```

  with, in `build_container.go` and `build_service.go`:

```go
func (s *service) buildContainer(_ context.Context, state *buildState) error {
	return applyContainer(state.req.Doc.Deployment.Container, state.req.Spec)
}
```

```go
func (s *service) buildService(_ context.Context, state *buildState) error {
	applyService(state.req.Doc.Deployment.Service, state.req.Spec)
	return nil
}
```

  (each file imports `"context"`). The template builders `buildHealthcheck`, `buildInit` and
  `buildResources` read the container and resources blocks through nil checks, so template builds
  stay as they were.

  *`specserviceimpl/build_settings.go`* - an import document's collection entries carry the id
  export wrote. Add:

```go
// entryBody is one entry of a collection block as the entity decodes it: an
// export's entry carries the id it was exported under, which is not a field of
// the setting.
func entryBody(state *buildState, entry any) any {
	body, ok := entry.(map[string]any)
	if !ok || !state.req.Import {
		return entry
	}
	body = maps.Clone(body)
	delete(body, specmodel.CollectionEntryIDKey)
	return body
}
```

  and in `buildSecrets` and `buildConfigFiles` decode `entryBody(state, entries[name])` in place of
  `entries[name]`.

  *`specserviceimpl/build_deployment.go`* - `buildSource` refuses a method other than an image only
  for a template:

```go
	if !state.req.Import && (source.ActiveMethod != base.DeploymentMethodImage || source.ImageSource == nil ||
		source.ImageSource.Image == "") {
		return invalidBlock(block, "an image to deploy is required")
	}
```

  *`specserviceimpl/build_test.go`* - the registry covers both lists:

```go
func TestBuilderRegistryCoversEveryBuildableBlock(t *testing.T) {
	registered := slices.Sorted(maps.Keys((&service{}).builders()))
	all := append(slices.Clone(specmodel.BuildableBlocks), specmodel.ImportOnlyBlocks...)
	assert.Equal(t, slices.Sorted(slices.Values(all)), registered)
}
```

- [ ] **Step 4: Run** the whole spec package - expect PASS. If the `repoSource` fields in
  `importDocYAML` do not match `entity.AppDeploymentSettings`' JSON names, correct them from that
  type (`grep -n 'json:"' hivepaas_app/entity/app_deployment_settings.go`) - `decodeBlock` refuses
  unknown fields, which is what would say so.
- [ ] **Step 5: Lint, commit** `feat(spec): build an exported document in import mode`.

---

### Task 7: Record it, and verify

- [ ] **Step 1:** In the import spec §9, after "**`container.image` is never written to a service.**",
  add a bullet: "**Import is a mode of `BuildApp`.** `BuildAppReq.Import` swaps `CheckBuildable`
  for `CheckImportable` and `PresentBlocks` for `ImportBlocks`; `deployment.container` and
  `deployment.service` are built only in it. What the settings screens keep for HivePaaS is kept
  here too: the labels `ApplyUserLabels` protects, the derived placement constraints, the network
  an app was created on, HivePaaS's log driver when none is named." In §14, item 3 names
  `docs/superpowers/plans/2026-09-24-spec-import-builders.md` and notes that app-scope setting types
  beyond kind, env vars, secrets, config files, routing and source follow in their own plan.
- [ ] **Step 2:** `go build ./... && golangci-lint run ./... && go test ./...` - all pass.
- [ ] **Step 3:** Commit `docs(spec): record import mode of the builder`, and report.
