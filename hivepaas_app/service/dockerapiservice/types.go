// Package dockerapiservice is the backend's side of giving apps the Docker API
// through the proxy each node's agent runs (pkg/dockerproxy): it turns
// app-docker-api settings into the proxy's policies, and asks the agents to act
// on a change now rather than at their next tick.
package dockerapiservice

const (
	// NetworkPrefix names an app's own network, which its children join.
	NetworkPrefix = "hp-dapi-"
	// NetworkLabel marks an app's own network, with the app's id. It is not the
	// proxy's owner label, so that the app can use the network but not remove it.
	NetworkLabel = "hivepaas.dockerApi.network"

	// SocketVolumePrefix names the volume an app's socket lives in, on every node.
	SocketVolumePrefix = "hp-dapi-sock-"
	// SocketVolumeLabel marks a socket volume, with the app's id. It is not the
	// owner label either: the app must not see or remove the volume its own
	// socket lives in.
	SocketVolumeLabel = "hivepaas.dockerApi.socket"

	// SocketMountTarget is where an app's socket volume is mounted: the directory
	// of dockerproxy.SocketPath.
	SocketMountTarget = "/var/run/hivepaas"

	// HostSocketPath is the node's own Docker socket, which host mode binds at
	// the same path in the app: where apps that need it look by default.
	HostSocketPath = "/var/run/docker.sock"

	// DefaultContainers, DefaultMemory and DefaultNanoCPUs are an app's limits
	// when its setting names none. DefaultCPUs is DefaultNanoCPUs as the setting
	// writes it.
	DefaultContainers = 5
	DefaultMemory     = 1 << 30
	DefaultNanoCPUs   = 1_000_000_000
	DefaultCPUs       = 1.0
)

// NetworkName is the name of an app's own network.
func NetworkName(appID string) string {
	return NetworkPrefix + appID
}

// SocketVolumeName is the name of the volume an app's socket lives in.
func SocketVolumeName(appID string) string {
	return SocketVolumePrefix + appID
}
