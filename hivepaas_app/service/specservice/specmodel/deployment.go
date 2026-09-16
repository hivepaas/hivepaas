package specmodel

import (
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/services/docker"
)

// Deployment is everything needed to recreate an app's running form: the
// app-deployment setting, plus the parts of the Swarm service HivePaaS treats
// as configuration.
//
// Those parts do not live in the settings table at all. Volume mounts,
// networks, resource limits, replicas, restart policy and container security
// options are read from the Swarm service - which is why a spec built from
// settings alone yields an app with the right environment variables and no
// storage.
//
// The blocks are sub-blocks rather than one flat struct for two reasons. The
// app-deployment setting has a Version and a Migrate; the Swarm-derived part
// has neither, and a future change to a Swarm field needs somewhere to hang a
// migration that is not the setting's version. And Command and WorkingDir exist
// in both, so Source carries the merged value and Container does not repeat it.
//
// The whole struct is optional. An app that has never been deployed has an
// empty ServiceID and no service to inspect - two of five user apps in a real
// development installation are in that state.
type Deployment struct {
	// Source is the assembled app-deployment setting.
	Source    map[string]any `yaml:"source,omitempty"`
	Container *Container     `yaml:"container,omitempty"`
	Resources *Resources     `yaml:"resources,omitempty"`
	Storage   *Storage       `yaml:"storage,omitempty"`
	Networks  *Networks      `yaml:"networks,omitempty"`
	Service   *Service       `yaml:"service,omitempty"`
}

// Container mirrors the writable half of a Swarm ContainerSpec.
//
// Command and WorkingDir are absent on purpose: they belong to Source, which
// carries the value the app-deployment setting and the service agree on.
type Container struct {
	ServiceLabels   map[string]string  `yaml:"serviceLabels,omitempty"`
	ContainerLabels map[string]string  `yaml:"containerLabels,omitempty"`
	Image           string             `yaml:"image,omitempty"`
	Hostname        string             `yaml:"hostname,omitempty"`
	User            string             `yaml:"user,omitempty"`
	Groups          []string           `yaml:"groups,omitempty"`
	StopSignal      string             `yaml:"stopSignal,omitempty"`
	TTY             bool               `yaml:"tty,omitempty"`
	Init            *bool              `yaml:"init,omitempty"`
	OpenStdin       bool               `yaml:"openStdin,omitempty"`
	ReadOnly        bool               `yaml:"readOnly,omitempty"`
	StopGracePeriod *timeutil.Duration `yaml:"stopGracePeriod,omitempty"`
	Privileges      *Privileges        `yaml:"privileges,omitempty"`
	Healthcheck     *Healthcheck       `yaml:"healthcheck,omitempty"`
	RestartPolicy   *RestartPolicy     `yaml:"restartPolicy,omitempty"`
	LogDriver       *LogDriver         `yaml:"logDriver,omitempty"`
}

type Privileges struct {
	SELinuxContext  *SELinuxContext `yaml:"seLinuxContext,omitempty"`
	Seccomp         *SeccompOpts    `yaml:"seccomp,omitempty"`
	AppArmor        *AppArmorOpts   `yaml:"appArmor,omitempty"`
	NoNewPrivileges bool            `yaml:"noNewPrivileges,omitempty"`
}

type SELinuxContext struct {
	Disable bool   `yaml:"disable,omitempty"`
	User    string `yaml:"user,omitempty"`
	Role    string `yaml:"role,omitempty"`
	Type    string `yaml:"type,omitempty"`
	Level   string `yaml:"level,omitempty"`
}

type SeccompOpts struct {
	Mode    swarm.SeccompMode `yaml:"mode,omitempty"`
	Profile string            `yaml:"profile,omitempty"`
}

type AppArmorOpts struct {
	Mode swarm.AppArmorMode `yaml:"mode,omitempty"`
}

type Healthcheck struct {
	Enabled bool                   `yaml:"enabled,omitempty"`
	Mode    docker.HealthcheckMode `yaml:"mode,omitempty"`
	Command string                 `yaml:"command,omitempty"`

	// Zero means inherit the image's own healthcheck.
	Interval      timeutil.Duration `yaml:"interval,omitempty"`
	Timeout       timeutil.Duration `yaml:"timeout,omitempty"`
	StartPeriod   timeutil.Duration `yaml:"startPeriod,omitempty"`
	StartInterval timeutil.Duration `yaml:"startInterval,omitempty"`
	Retries       int               `yaml:"retries,omitempty"`
}

type RestartPolicy struct {
	Condition   swarm.RestartPolicyCondition `yaml:"condition,omitempty"`
	Delay       *timeutil.Duration           `yaml:"delay,omitempty"`
	MaxAttempts *uint64                      `yaml:"maxAttempts,omitempty"`
	Window      *timeutil.Duration           `yaml:"window,omitempty"`
}

type LogDriver struct {
	Name    string            `yaml:"name,omitempty"`
	Options map[string]string `yaml:"options,omitempty"`
}

// Resources mirrors a Swarm TaskSpec's Resources plus the container limits
// HivePaaS exposes beside them.
type Resources struct {
	Reservations *ResourceReservations `yaml:"reservations,omitempty"`
	Limits       *ResourceLimits       `yaml:"limits,omitempty"`
	Memory       *Memory               `yaml:"memory,omitempty"`
	Capabilities *Capabilities         `yaml:"capabilities,omitempty"`
}

type ResourceReservations struct {
	CPUs             float64            `yaml:"cpus,omitempty"`
	Memory           unit.DataSize      `yaml:"memory,omitempty"`
	GenericResources []*GenericResource `yaml:"genericResources,omitempty"`
}

type GenericResource struct {
	Kind  string `yaml:"kind"`
	Value string `yaml:"value"`
}

type ResourceLimits struct {
	CPUs   float64       `yaml:"cpus,omitempty"`
	Memory unit.DataSize `yaml:"memory,omitempty"`
	Pids   int64         `yaml:"pids,omitempty"`
}

type Memory struct {
	Swap       *unit.DataSize `yaml:"swap,omitempty"`
	Swappiness *int64         `yaml:"swappiness,omitempty"`
	ShmSize    *unit.DataSize `yaml:"shmSize,omitempty"`
}

type Capabilities struct {
	Ulimits        []*Ulimit         `yaml:"ulimits,omitempty"`
	CapabilityAdd  []string          `yaml:"capabilityAdd,omitempty"`
	CapabilityDrop []string          `yaml:"capabilityDrop,omitempty"`
	EnableGPU      bool              `yaml:"enableGPU,omitempty"`
	OomScoreAdj    int64             `yaml:"oomScoreAdj,omitempty"`
	Sysctls        map[string]string `yaml:"sysctls,omitempty"`
}

type Ulimit struct {
	Name string `yaml:"name"`
	Hard int64  `yaml:"hard"`
	Soft int64  `yaml:"soft"`
}

// Storage keys mounts by their target path rather than by list position.
//
// A positional identity would be adequate while a spec carries only
// configuration - a reordered mount list re-imports the same either way. It
// stops being adequate as soon as anything outside the spec points at a mount,
// which is what a snapshot does when it pins volume data to one. Data pinned to
// mounts[0], a reorder, and a restore is data attached to the wrong volume,
// with no error and no warning.
//
// Two mounts at one target is meaningless, which makes the target a key - but
// nothing upstream validates it, so mapSwarmService checks and refuses.
type Storage struct {
	Mounts map[string]Mount `yaml:"mounts,omitempty"`
}

// Mount carries no Target: it is the map key. It carries no Key either, since
// that was a sha256 of the other three fields.
type Mount struct {
	Type           mount.Type        `yaml:"type"`
	Source         string            `yaml:"source,omitempty"`
	ReadOnly       bool              `yaml:"readOnly,omitempty"`
	Consistency    mount.Consistency `yaml:"consistency,omitempty"`
	BindOptions    *BindOptions      `yaml:"bindOptions,omitempty"`
	VolumeOptions  *VolumeOptions    `yaml:"volumeOptions,omitempty"`
	ClusterOptions *VolumeOptions    `yaml:"clusterOptions,omitempty"`
	TmpfsOptions   *TmpfsOptions     `yaml:"tmpfsOptions,omitempty"`
}

type BindOptions struct {
	Propagation            mount.Propagation `yaml:"propagation,omitempty"`
	NonRecursive           bool              `yaml:"nonRecursive,omitempty"`
	CreateMountpoint       bool              `yaml:"createMountpoint,omitempty"`
	ReadOnlyNonRecursive   bool              `yaml:"readOnlyNonRecursive,omitempty"`
	ReadOnlyForceRecursive bool              `yaml:"readOnlyForceRecursive,omitempty"`
}

type VolumeOptions struct {
	Subpath      string            `yaml:"subpath,omitempty"`
	NoCopy       bool              `yaml:"noCopy,omitempty"`
	Labels       map[string]string `yaml:"labels,omitempty"`
	DriverConfig *VolumeDriver     `yaml:"driverConfig,omitempty"`
}

type VolumeDriver struct {
	Name    string            `yaml:"name,omitempty"`
	Options map[string]string `yaml:"options,omitempty"`
}

type TmpfsOptions struct {
	Size    unit.DataSize     `yaml:"size,omitempty"`
	Mode    fileutil.FileMode `yaml:"mode,omitempty"`
	Options [][]string        `yaml:"options,omitempty"`
}

type Networks struct {
	Attachments      []*NetworkAttachment `yaml:"attachments,omitempty"`
	HostsFileEntries []*HostsFileEntry    `yaml:"hostsFileEntries,omitempty"`
	DNSConfig        *DNSConfig           `yaml:"dnsConfig,omitempty"`
	EndpointSpec     *EndpointSpec        `yaml:"endpointSpec,omitempty"`
}

// NetworkAttachment carries the network's name. The Swarm spec stores the id in
// Networks[].Target, which means nothing on another installation and nothing to
// a reader.
type NetworkAttachment struct {
	Name    string   `yaml:"name"`
	Aliases []string `yaml:"aliases,omitempty"`
}

type HostsFileEntry struct {
	Address   string   `yaml:"address,omitempty"`
	Hostnames []string `yaml:"hostnames,omitempty"`
}

type DNSConfig struct {
	Nameservers []string `yaml:"nameservers,omitempty"`
	Search      []string `yaml:"search,omitempty"`
	Options     []string `yaml:"options,omitempty"`
}

type EndpointSpec struct {
	Mode  swarm.ResolutionMode `yaml:"mode,omitempty"`
	Ports []*PortConfig        `yaml:"ports,omitempty"`
}

// PortConfig's Published is a reservation, recorded in res_links with dst_type
// "port". It travels with the spec and may collide on import.
type PortConfig struct {
	Target      uint32                      `yaml:"target,omitempty"`
	Published   uint32                      `yaml:"published,omitempty"`
	Protocol    network.IPProtocol          `yaml:"protocol,omitempty"`
	PublishMode swarm.PortConfigPublishMode `yaml:"publishMode,omitempty"`
}

type Service struct {
	ModeSpec  *ServiceModeSpec `yaml:"modeSpec,omitempty"`
	Placement *Placement       `yaml:"placement,omitempty"`
}

type ServiceModeSpec struct {
	Mode                docker.ServiceMode `yaml:"mode,omitempty"`
	ServiceReplicas     *uint64            `yaml:"serviceReplicas,omitempty"`
	JobMaxConcurrent    *uint64            `yaml:"jobMaxConcurrent,omitempty"`
	JobTotalCompletions *uint64            `yaml:"jobTotalCompletions,omitempty"`
}

// Placement carries no Platforms: Docker detects the platform at runtime, so it
// is not something a declarative document says.
type Placement struct {
	Constraints []string               `yaml:"constraints,omitempty"`
	Preferences []*PlacementPreference `yaml:"preferences,omitempty"`
}

type PlacementPreference struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}
