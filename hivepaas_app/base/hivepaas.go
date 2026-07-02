package base

const (
	HivepaasAppServiceName = "hivepaas_app"
	HivepaasAppKey         = "hivepaas_app"

	HivepaasWorkerServiceName = "hivepaas_worker"
	HivepaasWorkerKey         = "hivepaas_worker"

	HivepaasDbServiceName = "hivepaas_db"
	HivepaasDbAppKey      = "hivepaas_db"

	HivepaasCacheServiceName = "hivepaas_redis"
	HivepaasCacheAppKey      = "hivepaas_redis"

	HivepaasTraefikServiceName = "hivepaas_traefik"
	HivepaasTraefikAppKey      = "hivepaas_traefik"

	HivepaasUpdaterServiceName = "hivepaas_updater"
	HivepaasUpdaterKey         = "hivepaas_updater"

	HivepaasDockerProxyServiceName = "hivepaas_docker_proxy"
	HivepaasDockerProxyKey         = "hivepaas_docker_proxy"

	HivepaasAgentServiceName = "hivepaas_agent"
	HivepaasAgentKey         = "hivepaas_agent"
)

const (
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
