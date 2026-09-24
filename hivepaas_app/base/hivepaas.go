package base

const (
	HivepaasAppServiceName = "hivepaas_app"
	HivepaasAppKey         = "app"

	HivepaasWorkerServiceName = "hivepaas_worker"
	HivepaasWorkerKey         = "worker"

	HivepaasDbServiceName = "hivepaas_db"
	HivepaasDbKey         = "db"

	HivepaasCacheServiceName = "hivepaas_redis"
	HivepaasCacheKey         = "redis"

	HivepaasTraefikServiceName = "hivepaas_traefik"
	HivepaasTraefikKey         = "traefik"

	HivepaasUpdaterServiceName = "hivepaas_updater"
	HivepaasUpdaterKey         = "updater"

	HivepaasDockerProxyServiceName = "hivepaas_docker_proxy"
	HivepaasDockerProxyKey         = "docker_proxy"

	HivepaasAgentServiceName = "hivepaas_agent"
	HivepaasAgentKey         = "agent"

	// The logging stack is two apps the app provisions in its hidden project when
	// logging is switched on, not services of the stack file like everything
	// above. These are their app keys. Either being absent means HivePaaS does not
	// run that part, not that something is broken.
	HivepaasVictoriaLogsKey = "victoria-logs"
	HivepaasVlagentKey      = "vlagent"

	// The registry is created by the app when an operator switches it on, not by
	// the stack file, so it carries no stack prefix either. Its absence means the
	// registry was never switched on, not that something is broken.
	HivepaasRegistryKey = "registry"
)

const (
	HivepaasScope       = "hivepaas"
	HivepaasProjectName = "HivePaaS"
	HivepaasProjectKey  = "hivepaas"
)

var (
	UnallowedProjectKeys = []string{HivepaasProjectKey}
)

const (
	NetworkGlobalRouting = "hivepaas_net"
	NetworkDockerProxy   = "hivepaas_docker_proxy_net"
	NetworkHivepaasLocal = "hivepaas_local_net"
)

// OomScoreAdjSystemAddon is the kernel OOM priority of the services an admin adds
// to HivePaaS - the log collector, the log backend, the registry. When memory
// runs out a user app (0) is killed first, and these go before the core stack
// (-500, set by the installer) because running apps do not depend on them.
//
// Only a service with a memory limit gets it: a protected service without one
// that leaks would have the kernel kill every user app to feed it.
const OomScoreAdjSystemAddon = -300
