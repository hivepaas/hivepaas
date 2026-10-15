package imagescanservice

import (
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/obi"
)

// ComponentIDPrefix starts the ID HivePaaS's components are kept under, where
// an app's results are kept by the app's ID.
const ComponentIDPrefix = "hivepaas:"

// Component is a part of HivePaaS whose images are scanned with the apps':
// what it runs is found by one of Service, SystemAppKey and Container.
type Component struct {
	Key  string
	Name string
	// Service is the name of its swarm service in HivePaaS's stack.
	Service string
	// SystemAppKey is the key of the system app it is, when it is one.
	SystemAppKey string
	// Container is the name of the container each node's agent runs.
	Container string
}

// ID is the key its results are kept under.
func (c Component) ID() string { return ComponentIDPrefix + c.Key }

// Components are HivePaaS's parts, in the order they are shown. The worker
// and the updater run the app's image, and are left out with it.
var Components = []Component{
	{Key: "app", Name: "HivePaaS", Service: base.HivepaasAppServiceName},
	{Key: "agent", Name: "Agent", Service: base.HivepaasAgentServiceName},
	{Key: "db", Name: "Database", Service: base.HivepaasDbServiceName},
	{Key: "redis", Name: "Redis", Service: base.HivepaasCacheServiceName},
	{Key: "traefik", Name: "Traefik", Service: base.HivepaasTraefikServiceName},
	{Key: base.HivepaasRegistryKey, Name: "Registry", SystemAppKey: base.HivepaasRegistryKey},
	{Key: base.HivepaasVictoriaLogsKey, Name: "Logging backend", SystemAppKey: base.HivepaasVictoriaLogsKey},
	{Key: base.HivepaasVlagentKey, Name: "Log collector", SystemAppKey: base.HivepaasVlagentKey},
	{Key: base.HivepaasOBIKey, Name: "OBI", Container: obi.ContainerName},
}

// ComponentByKey is the component of a key; false when there is none.
func ComponentByKey(key string) (Component, bool) {
	for _, c := range Components {
		if c.Key == key {
			return c, true
		}
	}
	return Component{}, false
}

// IsComponentID says an ID is a component's, not an app's.
func IsComponentID(id string) bool { return strings.HasPrefix(id, ComponentIDPrefix) }

// ComponentIgnores are the ignores that apply to HivePaaS's components: the
// global ones, inherited by the apps or not - a component is of no project.
func ComponentIgnores(settings []*entity.Setting) []*Ignore {
	var out []*Ignore
	for _, setting := range settings {
		if setting.Type != base.SettingTypeVulnerabilityIgnore || setting.Status != base.SettingStatusActive ||
			setting.Scope != base.ObjectScopeGlobal {
			continue
		}
		data, err := setting.AsVulnerabilityIgnore()
		if err != nil {
			continue
		}
		out = append(out, &Ignore{Setting: setting, Data: data})
	}
	return out
}
