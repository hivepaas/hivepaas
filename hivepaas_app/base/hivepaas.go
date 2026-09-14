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

	// The logging stack is created by the app when logging is switched on, not
	// by the stack file like everything above, which is why these two are spelled
	// with hyphens and carry no stack prefix. Either being absent means logging
	// is off, not that something is broken.
	HivepaasVictoriaLogsServiceName = "hivepaas-victoria-logs"
	HivepaasVictoriaLogsKey         = "victoria-logs"

	HivepaasVlagentServiceName = "hivepaas-vlagent"
	HivepaasVlagentKey         = "vlagent"
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
	// NetworkLogging is the overlay the collector and the backend talk over.
	// Only the logging stack joins it.
	NetworkLogging = "hivepaas_logging_net"
)
