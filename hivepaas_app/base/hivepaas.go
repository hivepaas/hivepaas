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
