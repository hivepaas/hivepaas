package base

const (
	HivepaasAppServiceName = "hivepaas_app"
	HivepaasAppGlobalKey   = "hivepaas_app"

	HivepaasWorkerServiceName = "hivepaas_worker"
	HivepaasWorkerGlobalKey   = "hivepaas_worker"

	HivepaasDbServiceName = "hivepaas_db"
	HivepaasDbGlobalKey   = "hivepaas_db"

	HivepaasCacheServiceName = "hivepaas_redis"
	HivepaasCacheGlobalKey   = "hivepaas_redis"

	HivepaasTraefikServiceName = "hivepaas_traefik"
	HivepaasTraefikGlobalKey   = "hivepaas_traefik"

	HivepaasUpdaterServiceName = "hivepaas_updater"
	HivepaasUpdaterGlobalKey   = "hivepaas_updater"

	HivepaasDockerProxyServiceName = "hivepaas_docker_proxy"
	HivepaasDockerProxyGlobalKey   = "hivepaas_docker_proxy"

	HivepaasAgentServiceName = "hivepaas_agent"
	HivepaasAgentGlobalKey   = "hivepaas_agent"
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
