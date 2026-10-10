package composeserviceimpl

import (
	"maps"
	"net/netip"
	"slices"
	"strconv"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/services/docker"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

// app is the app a service becomes, and what the review shows of it.
func (c *converter) app(name string, svc types.ServiceConfig) (*specmodel.AppDoc, *composeservice.ServiceView) {
	path, key := c.appPath(name), c.keys[name]
	doc := &specmodel.AppDoc{App: key, Name: name, Status: string(base.AppStatusActive)}
	view := &composeservice.ServiceView{Name: name, App: key, Build: svc.Build != nil, Aliases: c.aliases[name],
		Dropped: c.r.unsupported[name], Existing: c.existing[name], UseExisting: c.used[name] != ""}
	if existing := c.existingApp(name); existing != nil {
		c.add(path, "", composeservice.CodeAppUsed, map[string]any{detailApp: key},
			"the env's app is used, as it is: the others reach it by its name")
		return existing, view
	}

	image := c.plain(path, "image", svc.Image)
	if choice := c.req.Services[name]; image == "" && choice != nil {
		image = strings.TrimSpace(choice.Image)
	}
	view.Image = image
	if image == "" {
		view.Skipped, view.Reason = true, "the file builds its image"
		c.add(path, specmodel.SeveritySkipped, composeservice.CodeNoImage, nil,
			"not created: HivePaaS builds from a repository - give it an image")
		return doc, view
	}
	if svc.Build != nil {
		c.add(path, specmodel.SeverityWarning, composeservice.CodeBuildIgnored, map[string]any{detailImage: image},
			"the image is deployed, and the file's build left out")
	}
	if len(view.Dropped) > 0 {
		c.add(path, specmodel.SeverityWarning, composeservice.CodeNotSupported, map[string]any{"fields": view.Dropped},
			"left out: Swarm services have none of them, or HivePaaS does not set it")
	}

	settings, mounts := map[string]any{}, map[string]any{}
	deployment := &specmodel.Deployment{
		Source:    c.source(path, image, svc),
		Container: c.container(path, name, svc),
		Resources: c.resources(path, svc),
		Service:   c.serviceBlock(path, name, svc, view),
		Networks:  c.networks(path, name, svc),
	}
	deployment.Storage = c.storage(path, name, svc, view, mounts)
	c.ports(path, name, svc, view, deployment.Networks, settings)
	secrets := map[string]any{}
	envVars, kept := c.envVars(path, name, svc, secrets)
	if envVars != nil {
		settings["envVars"] = envVars
	}
	if len(secrets) > 0 {
		settings[blockSecrets] = secrets
	}
	view.Secrets = kept
	c.fileMounts(svc, mounts)
	c.writableForFiles(path, deployment.Storage, mounts, view)
	if len(mounts) > 0 {
		settings["settingMounts"] = mounts
	}
	doc.Deployment = deployment
	if len(settings) > 0 {
		doc.Settings = settings
	}
	c.notes(path, name, svc, image)
	return doc, view
}

// plain is a value of the file other than an environment's: a secret
// variable in it is written as its value, which a note says.
func (c *converter) plain(path, field, s string) string {
	out, _ := c.r.markers.replacePlain(s, c.r.values)
	out, names := c.r.markers.replace(out, func(name string) string { return c.r.values[name] })
	for _, name := range names {
		c.add(path, "", composeservice.CodeSecretWritten, map[string]any{"variable": name, detailField: field},
			"written as its value: only an environment refers to a secret")
	}
	return out
}

func (c *converter) source(path, image string, svc types.ServiceConfig) map[string]any {
	source := map[string]any{
		"activeMethod":           string(base.DeploymentMethodImage),
		"imageSource":            map[string]any{detailImage: image},
		specmodel.SettingMetaKey: map[string]any{"version": entity.CurrentAppDeploymentSettingsVersion},
	}
	if line := c.commandLine(path, "entrypoint", svc.Entrypoint); line != "" {
		source["entrypoint"] = line
	}
	if line := c.commandLine(path, "command", svc.Command); line != "" {
		source["command"] = line
	}
	if dir := c.plain(path, "working_dir", svc.WorkingDir); dir != "" {
		source["workingDir"] = dir
	}
	return source
}

// commandLine is an argv as the deployment settings write it: quoted, so that
// a deployment splits it back. An empty one - clearing the image's - is not
// something HivePaaS can say.
func (c *converter) commandLine(path, field string, argv types.ShellCommand) string {
	if argv == nil {
		return ""
	}
	if len(argv) == 0 || len(argv) == 1 && argv[0] == "" {
		c.add(path, specmodel.SeverityFixable, composeservice.CodeValueDropped, map[string]any{detailField: field},
			"an empty one is left out: the image's is kept")
		return ""
	}
	words := make([]string, len(argv))
	for i, word := range argv {
		words[i] = c.plain(path, field, word)
	}
	return dockerhelper.CommandLine(words)
}

func (c *converter) container(path, name string, svc types.ServiceConfig) *specmodel.Container {
	out := &specmodel.Container{
		// The import clears what it is not given: the hostname is the app's key.
		Hostname:        c.plain(path, "hostname", svc.Hostname),
		User:            c.plain(path, "user", svc.User),
		Groups:          svc.GroupAdd,
		StopSignal:      svc.StopSignal,
		TTY:             svc.Tty,
		Init:            svc.Init,
		OpenStdin:       svc.StdinOpen,
		ReadOnly:        svc.ReadOnly,
		Privileges:      c.privileges(path, svc.SecurityOpt),
		Healthcheck:     c.healthcheck(path, svc.HealthCheck),
		RestartPolicy:   restartPolicy(svc),
		LogDriver:       logDriver(svc),
		ContainerLabels: c.labels(path, "labels", svc.Labels),
	}
	if out.Hostname == "" {
		out.Hostname = c.keys[name]
	}
	if svc.StopGracePeriod != nil {
		out.StopGracePeriod = new(timeutil.Duration(*svc.StopGracePeriod))
	}
	if svc.Deploy != nil {
		out.ServiceLabels = c.labels(path, "deploy.labels", svc.Deploy.Labels)
	}
	return out
}

// droppedLabelPrefixes are labels HivePaaS keeps for its own use: an app's are
// dropped, as it drops them anywhere else.
var droppedLabelPrefixes = []string{"traefik.", "hivepaas.", "com.docker.stack.", "com.docker.compose."}

func (c *converter) labels(path, field string, labels types.Labels) map[string]string {
	out := map[string]string{}
	var dropped []string
	for key, value := range labels {
		if slices.ContainsFunc(droppedLabelPrefixes, func(prefix string) bool { return strings.HasPrefix(key, prefix) }) {
			dropped = append(dropped, key)
			continue
		}
		out[key] = c.plain(path, field, value)
	}
	if len(dropped) > 0 {
		slices.Sort(dropped)
		c.add(path, "", composeservice.CodeLabelsDropped, map[string]any{detailField: field, "labels": dropped},
			"left out: HivePaaS keeps these for itself, and routes by the app's domains")
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// privileges reads security_opt: what Swarm has a field for.
func (c *converter) privileges(path string, opts []string) *specmodel.Privileges {
	out := &specmodel.Privileges{}
	var dropped []string
	for _, opt := range opts {
		key, value := cutOpt(opt)
		switch {
		case key == "no-new-privileges":
			out.NoNewPrivileges = value == "" || value == "true"
		case key == "seccomp" && value == "unconfined":
			out.Seccomp = &specmodel.SeccompOpts{Mode: swarm.SeccompModeUnconfined}
		case key == "apparmor" && value == "unconfined":
			out.AppArmor = &specmodel.AppArmorOpts{Mode: swarm.AppArmorModeDisabled}
		case key == "label" && seLinux(out, value):
		default:
			dropped = append(dropped, opt)
		}
	}
	if len(dropped) > 0 {
		c.add(path, specmodel.SeverityFixable, composeservice.CodeValueDropped,
			map[string]any{detailField: "security_opt", "values": dropped}, "left out: Swarm has no field for it")
	}
	if !out.NoNewPrivileges && out.Seccomp == nil && out.AppArmor == nil && out.SELinuxContext == nil {
		return nil
	}
	return out
}

// cutOpt reads `key:value` or `key=value`.
func cutOpt(opt string) (string, string) {
	if i := strings.IndexAny(opt, ":="); i >= 0 {
		return opt[:i], opt[i+1:]
	}
	return opt, ""
}

// seLinux reads a label option into a SELinux context; false for one it
// does not know.
func seLinux(p *specmodel.Privileges, value string) bool {
	if p.SELinuxContext == nil {
		p.SELinuxContext = &specmodel.SELinuxContext{}
	}
	ctx := p.SELinuxContext
	field, v := cutOpt(value)
	switch field {
	case "disable":
		ctx.Disable = true
	case "user":
		ctx.User = v
	case "role":
		ctx.Role = v
	case "type":
		ctx.Type = v
	case "level":
		ctx.Level = v
	default:
		return false
	}
	return true
}

// healthcheck reads a healthcheck as the container settings write one: CMD's
// argv quoted, CMD-SHELL's string whole, NONE kept, and no test the image's.
func (c *converter) healthcheck(path string, hc *types.HealthCheckConfig) *specmodel.Healthcheck {
	if hc == nil {
		return nil
	}
	if hc.Disable || len(hc.Test) > 0 && hc.Test[0] == string(docker.HealthcheckModeNone) {
		return &specmodel.Healthcheck{Mode: docker.HealthcheckModeNone}
	}
	out := &specmodel.Healthcheck{Enabled: true}
	if len(hc.Test) > 0 {
		test := make([]string, len(hc.Test)-1)
		for i, word := range hc.Test[1:] {
			test[i] = c.plain(path, "healthcheck", word)
		}
		switch docker.HealthcheckMode(hc.Test[0]) {
		case docker.HealthcheckModeCmd:
			out.Mode, out.Command = docker.HealthcheckModeCmd, dockerhelper.CommandLine(test)
		case docker.HealthcheckModeCmdShell, docker.HealthcheckModeInherit, docker.HealthcheckModeNone:
			out.Mode, out.Command = docker.HealthcheckModeCmdShell, strings.Join(test, " ")
		default:
			out.Mode, out.Command = docker.HealthcheckModeCmdShell, strings.Join(test, " ")
		}
	}
	out.Interval, out.Timeout = duration(hc.Interval), duration(hc.Timeout)
	out.StartPeriod, out.StartInterval = duration(hc.StartPeriod), duration(hc.StartInterval)
	if hc.Retries != nil {
		out.Retries = int(*hc.Retries) //nolint:gosec // a count of retries
	}
	return out
}

func duration(d *types.Duration) timeutil.Duration {
	if d == nil {
		return 0
	}
	return timeutil.Duration(*d)
}

func durationRef(d *types.Duration) *timeutil.Duration {
	if d == nil {
		return nil
	}
	return new(timeutil.Duration(*d))
}

// restartPolicy is deploy.restart_policy, which Swarm reads, or else restart:
// `unless-stopped` is `any` there, as a stopped Swarm task is replaced anyway.
func restartPolicy(svc types.ServiceConfig) *specmodel.RestartPolicy {
	if svc.Deploy != nil && svc.Deploy.RestartPolicy != nil {
		rp := svc.Deploy.RestartPolicy
		return &specmodel.RestartPolicy{Condition: swarm.RestartPolicyCondition(rp.Condition),
			Delay: durationRef(rp.Delay), MaxAttempts: rp.MaxAttempts, Window: durationRef(rp.Window)}
	}
	policy, attempts, _ := strings.Cut(svc.Restart, ":")
	switch policy {
	case types.RestartPolicyNo:
		return &specmodel.RestartPolicy{Condition: swarm.RestartPolicyConditionNone}
	case types.RestartPolicyAlways, types.RestartPolicyUnlessStopped:
		return &specmodel.RestartPolicy{Condition: swarm.RestartPolicyConditionAny}
	case types.RestartPolicyOnFailure:
		out := &specmodel.RestartPolicy{Condition: swarm.RestartPolicyConditionOnFailure}
		if n, err := strconv.ParseUint(attempts, 10, 64); err == nil {
			out.MaxAttempts = &n
		}
		return out
	}
	return nil
}

func logDriver(svc types.ServiceConfig) *specmodel.LogDriver {
	switch {
	case svc.Logging != nil && svc.Logging.Driver != "":
		return &specmodel.LogDriver{Name: svc.Logging.Driver, Options: maps.Clone(svc.Logging.Options)}
	case svc.LogDriver != "":
		return &specmodel.LogDriver{Name: svc.LogDriver, Options: maps.Clone(svc.LogOpt)}
	}
	return nil
}

// resources are deploy.resources, or the older fields beside it, and what a
// container is granted beyond an ordinary one - which the caller may not.
func (c *converter) resources(path string, svc types.ServiceConfig) *specmodel.Resources {
	var limits, reservations *types.Resource
	if svc.Deploy != nil {
		limits, reservations = svc.Deploy.Resources.Limits, svc.Deploy.Resources.Reservations
	}
	gpus, allGPUs := gpusWanted(svc, reservations)
	out := &specmodel.Resources{
		Limits:       resourceLimits(limits, svc),
		Reservations: c.resourceReservations(path, reservations, svc),
		Memory:       memory(svc),
		Capabilities: c.capabilities(path, svc, gpus > 0),
	}
	if out.Capabilities != nil && out.Capabilities.EnableGPU {
		c.reserveGPUs(path, out, gpus, allGPUs)
	}
	if out.Limits == nil && out.Reservations == nil && out.Memory == nil && out.Capabilities == nil {
		return nil
	}
	return out
}

func resourceLimits(r *types.Resource, svc types.ServiceConfig) *specmodel.ResourceLimits {
	out := &specmodel.ResourceLimits{CPUs: float64(svc.CPUS), Memory: bytesOf(svc.MemLimit), Pids: svc.PidsLimit}
	if r != nil {
		if r.NanoCPUs > 0 {
			out.CPUs = float64(r.NanoCPUs)
		}
		if r.MemoryBytes > 0 {
			out.Memory = bytesOf(r.MemoryBytes)
		}
		if r.Pids > 0 {
			out.Pids = r.Pids
		}
	}
	if out.CPUs <= 0 && out.Memory <= 0 && out.Pids <= 0 {
		return nil
	}
	return out
}

func (c *converter) resourceReservations(
	path string, r *types.Resource, svc types.ServiceConfig,
) *specmodel.ResourceReservations {
	out := &specmodel.ResourceReservations{Memory: bytesOf(svc.MemReservation)}
	if r != nil {
		out.CPUs = float64(r.NanoCPUs)
		if r.MemoryBytes > 0 {
			out.Memory = bytesOf(r.MemoryBytes)
		}
		for _, generic := range r.GenericResources {
			spec := generic.DiscreteResourceSpec
			if spec == nil {
				continue
			}
			// GPUs reserved by hand are granted as much as `gpus`.
			if spec.Kind == docker.GenericResourceGPU && !c.req.MayWriteCluster {
				c.capabilitiesDropped(path)
				continue
			}
			out.GenericResources = append(out.GenericResources, &specmodel.GenericResource{
				Kind: spec.Kind, Value: strconv.FormatInt(spec.Value, 10),
			})
		}
	}
	if out.CPUs <= 0 && out.Memory <= 0 && len(out.GenericResources) == 0 {
		return nil
	}
	return out
}

func memory(svc types.ServiceConfig) *specmodel.Memory {
	out := &specmodel.Memory{}
	if svc.MemSwapLimit != 0 {
		out.Swap = new(bytesOf(svc.MemSwapLimit))
	}
	if svc.MemSwappiness > 0 {
		out.Swappiness = new(int64(svc.MemSwappiness))
	}
	if svc.ShmSize > 0 {
		out.ShmSize = new(bytesOf(svc.ShmSize))
	}
	if out.Swap == nil && out.Swappiness == nil && out.ShmSize == nil {
		return nil
	}
	return out
}

// capabilities are what a container is granted beyond an ordinary one. They
// take Write on the Cluster module: without it, the app is created without
// them.
func (c *converter) capabilities(path string, svc types.ServiceConfig, wantsGPU bool) *specmodel.Capabilities {
	out := &specmodel.Capabilities{
		CapabilityAdd: svc.CapAdd, CapabilityDrop: svc.CapDrop, OomScoreAdj: svc.OomScoreAdj,
		Sysctls: map[string]string(svc.Sysctls), EnableGPU: wantsGPU,
	}
	for _, name := range slices.Sorted(maps.Keys(svc.Ulimits)) {
		limit := svc.Ulimits[name]
		if limit == nil {
			continue
		}
		soft, hard := int64(limit.Soft), int64(limit.Hard)
		if limit.Single != 0 {
			soft, hard = int64(limit.Single), int64(limit.Single)
		}
		out.Ulimits = append(out.Ulimits, &specmodel.Ulimit{Name: name, Soft: soft, Hard: hard})
	}
	if len(out.CapabilityAdd) == 0 && len(out.CapabilityDrop) == 0 && out.OomScoreAdj == 0 &&
		len(out.Sysctls) == 0 && !out.EnableGPU && len(out.Ulimits) == 0 {
		return nil
	}
	if !c.req.MayWriteCluster {
		c.capabilitiesDropped(path)
		return nil
	}
	return out
}

// capabilitiesDropped says, once for the service, that what it is granted
// beyond an ordinary container is left out.
func (c *converter) capabilitiesDropped(path string) {
	if slices.ContainsFunc(c.resp.Issues[path], func(issue specmodel.Issue) bool {
		return issue.Code == composeservice.CodeCapabilityDropped
	}) {
		return
	}
	c.add(path, specmodel.SeverityFixable, composeservice.CodeCapabilityDropped, nil,
		"left out: granting capabilities, ulimits, sysctls or GPUs takes Write on the Cluster module")
}

// reserveGPUs reserves what the service asks of GPUs: one is enableGPU, more
// a count of them among the generic resources. Swarm reserves GPUs by count:
// every one a node has is reserved as one.
func (c *converter) reserveGPUs(path string, out *specmodel.Resources, gpus int, allGPUs bool) {
	if allGPUs {
		c.add(path, specmodel.SeverityFixable, composeservice.CodeValueDropped, map[string]any{detailField: "gpus"},
			"one GPU is reserved for `all`: swarm reserves GPUs by count")
	}
	if gpus <= 1 {
		return
	}
	out.Capabilities.EnableGPU = false
	if out.Reservations == nil {
		out.Reservations = &specmodel.ResourceReservations{}
	}
	out.Reservations.GenericResources = append(out.Reservations.GenericResources, &specmodel.GenericResource{
		Kind: docker.GenericResourceGPU, Value: strconv.Itoa(gpus),
	})
	capabilities := out.Capabilities
	if len(capabilities.CapabilityAdd) == 0 && len(capabilities.CapabilityDrop) == 0 && capabilities.OomScoreAdj == 0 &&
		len(capabilities.Sysctls) == 0 && len(capabilities.Ulimits) == 0 {
		out.Capabilities = nil
	}
}

// gpusWanted is how many GPUs the service asks for - by `gpus`, or by a
// device reservation for them - and whether one of them asks for all a node
// has.
func gpusWanted(svc types.ServiceConfig, reservations *types.Resource) (count int, all bool) {
	requests := slices.Clone(svc.Gpus)
	if reservations != nil {
		for _, device := range reservations.Devices {
			if slices.Contains(device.Capabilities, "gpu") {
				requests = append(requests, device)
			}
		}
	}
	for _, request := range requests {
		switch {
		case request.Count < 0:
			all = true
			count++
		case request.Count > 0:
			count += int(request.Count)
		case len(request.IDs) > 0:
			count += len(request.IDs)
		default:
			count++
		}
	}
	return count, all
}

func bytesOf(b types.UnitBytes) unit.DataSize {
	return unit.DataSize(b)
}

// serviceBlock is the service's mode and placement: replicated with its
// replicas, global, or a job - a service another waits on to complete.
func (c *converter) serviceBlock(
	path, name string, svc types.ServiceConfig, view *composeservice.ServiceView,
) *specmodel.Service {
	spec := &specmodel.ServiceModeSpec{Mode: docker.ServiceModeReplicated}
	mode := ""
	if svc.Deploy != nil {
		mode = svc.Deploy.Mode
	}
	switch docker.ServiceMode(mode) {
	case docker.ServiceModeGlobal, docker.ServiceModeReplicatedJob, docker.ServiceModeGlobalJob:
		spec.Mode = docker.ServiceMode(mode)
	case docker.ServiceModeReplicated:
		spec.ServiceReplicas = replicasOf(svc)
	default:
		spec.ServiceReplicas = replicasOf(svc)
	}
	if c.jobs[name] && spec.Mode == docker.ServiceModeReplicated {
		one := uint64(1)
		spec.Mode, spec.ServiceReplicas = docker.ServiceModeReplicatedJob, nil
		spec.JobMaxConcurrent, spec.JobTotalCompletions = &one, new(one)
	}
	view.Mode = string(spec.Mode)
	if spec.ServiceReplicas != nil {
		view.Replicas = *spec.ServiceReplicas
	}

	out := &specmodel.Service{ModeSpec: spec}
	if svc.Deploy != nil && (len(svc.Deploy.Placement.Constraints) > 0 || len(svc.Deploy.Placement.Preferences) > 0) {
		out.Placement = &specmodel.Placement{Constraints: svc.Deploy.Placement.Constraints}
		for _, pref := range svc.Deploy.Placement.Preferences {
			out.Placement.Preferences = append(out.Placement.Preferences,
				&specmodel.PlacementPreference{Name: "spread", Value: pref.Spread})
		}
		c.add(path, specmodel.SeverityWarning, composeservice.CodePlacement,
			map[string]any{"constraints": svc.Deploy.Placement.Constraints},
			"kept: they name nodes and labels of the cluster the file was written for")
	}
	return out
}

func replicasOf(svc types.ServiceConfig) *uint64 {
	replicas := uint64(1)
	switch {
	case svc.Deploy != nil && svc.Deploy.Replicas != nil && *svc.Deploy.Replicas >= 0:
		replicas = uint64(*svc.Deploy.Replicas)
	case svc.Scale != nil && *svc.Scale >= 0:
		replicas = uint64(*svc.Scale)
	}
	return &replicas
}

// networks attach the app to its env's network, by its key and the names the
// file reaches it by; and carry its hosts and DNS.
func (c *converter) networks(path, name string, svc types.ServiceConfig) *specmodel.Networks {
	out := &specmodel.Networks{Attachments: []*specmodel.NetworkAttachment{{
		Name: c.req.NetworkName, Aliases: append([]string{c.keys[name]}, c.aliases[name]...),
	}}}
	for _, host := range slices.Sorted(maps.Keys(svc.ExtraHosts)) {
		for _, address := range svc.ExtraHosts[host] {
			out.HostsFileEntries = append(out.HostsFileEntries,
				&specmodel.HostsFileEntry{Address: address, Hostnames: []string{host}})
		}
	}
	if len(svc.DNS) > 0 || len(svc.DNSSearch) > 0 || len(svc.DNSOpts) > 0 {
		dns := &specmodel.DNSConfig{Search: svc.DNSSearch, Options: svc.DNSOpts}
		for _, server := range svc.DNS {
			if _, err := netip.ParseAddr(server); err != nil {
				c.add(path, specmodel.SeverityFixable, composeservice.CodeValueDropped,
					map[string]any{detailField: "dns", "values": []string{server}}, "left out: not an address")
				continue
			}
			dns.Nameservers = append(dns.Nameservers, server)
		}
		out.DNSConfig = dns
	}
	c.networkWarnings(path, svc)
	return out
}

// networkWarnings say what of a service's networks is not kept: every app is
// on its env's one network.
func (c *converter) networkWarnings(path string, svc types.ServiceConfig) {
	for _, network := range slices.Sorted(maps.Keys(svc.Networks)) {
		var fields []string
		if config := svc.Networks[network]; config != nil {
			if config.Ipv4Address != "" || config.Ipv6Address != "" || len(config.LinkLocalIPs) > 0 {
				fields = append(fields, "addresses")
			}
			if config.MacAddress != "" || config.InterfaceName != "" || config.Priority != 0 || len(config.DriverOpts) > 0 {
				fields = append(fields, "options")
			}
		}
		if declared, ok := c.r.project.Networks[network]; ok && bool(declared.External) {
			fields = append(fields, "external")
		}
		if len(fields) > 0 {
			c.add(path, specmodel.SeverityWarning, composeservice.CodeNetwork,
				map[string]any{"network": network, "fields": fields},
				"not kept: every app is on its env's one network, and reaches the others there")
		}
	}
}

// notes say what the import does that the file may not expect.
func (c *converter) notes(path, name string, svc types.ServiceConfig, image string) {
	if aliases := c.aliases[name]; len(aliases) > 0 {
		c.add(path, "", composeservice.CodeAliasAdded, map[string]any{"aliases": aliases},
			"the app is reached by these names too, as in compose")
	}
	var waits []string
	for _, dep := range slices.Sorted(maps.Keys(svc.DependsOn)) {
		if svc.DependsOn[dep].Condition == types.ServiceConditionHealthy {
			waits = append(waits, dep)
		}
	}
	if len(waits) > 0 {
		c.add(path, "", composeservice.CodeStartOrder, map[string]any{detailServices: waits},
			"not waited for: the app is restarted until what it needs answers")
	}
	if c.jobs[name] {
		c.add(path, "", composeservice.CodeJob, nil, "run to completion on each deployment: another waits on it")
	}
	if !imageHasTag(image) {
		c.add(path, "", composeservice.CodeImageUnpinned, map[string]any{detailImage: image},
			"no tag: each deployment pulls the newest")
	}
}

// imageHasTag says whether an image names a tag other than latest, or a
// digest.
func imageHasTag(image string) bool {
	if strings.Contains(image, "@") {
		return true
	}
	name := image[strings.LastIndex(image, "/")+1:]
	_, tag, found := strings.Cut(name, ":")
	return found && tag != "" && tag != "latest"
}
