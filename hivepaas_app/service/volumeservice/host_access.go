package volumeservice

import (
	"path/filepath"
	"strings"
)

const (
	// DockerSocketPath is the daemon's socket on a node. An app that mounts it is
	// root on that node, and on a manager over the whole cluster.
	DockerSocketPath = "/var/run/docker.sock"

	// driverOptType, driverOptDevice and driverOptO describe, to docker's local
	// driver, what to mount and from where. A volume carrying them is a mount of
	// the host by another name.
	driverOptType = "type"
	// DriverOptDevice names what a local volume mounts, and is the one option a
	// caller may write a path of the node into.
	DriverOptDevice = "device"
	driverOptO      = "o"
)

// hostDriverOpts are the driver options that make a local volume reach a path
// of the node rather than docker's own storage.
var hostDriverOpts = []string{driverOptType, DriverOptDevice, driverOptO}

// ReachesDockerSocket reports whether a path of the node gives access to the
// docker socket: the socket itself, or a directory holding it - /var/run, /var,
// / and so on. Access to the Docker API is granted through an app's Docker API
// settings, where what it may do is written down; a mount is not that grant and
// never stands in for it.
func ReachesDockerSocket(path string) bool {
	if path == "" {
		return false
	}
	clean := filepath.Clean(path)
	if !strings.HasPrefix(clean, "/") {
		return false
	}
	if clean == DockerSocketPath {
		return true
	}
	// A directory at or above the socket's own holds it.
	return strings.HasPrefix(DockerSocketPath, strings.TrimSuffix(clean, "/")+"/")
}

// DriverOptsReachDockerSocket reports the same for a volume's driver options:
// the device they name is a path of the node.
func DriverOptsReachDockerSocket(driverOpts map[string]string) bool {
	return ReachesDockerSocket(driverOpts[DriverOptDevice])
}

// DriverOptsNameHostPath reports whether driver options describe a mount of the
// node's filesystem. HivePaaS writes these itself for a volume whose directory
// it chose, so the answer is about the description, not about who wrote it: the
// caller decides which of the two it is asking about.
func DriverOptsNameHostPath(driverOpts map[string]string) bool {
	for _, key := range hostDriverOpts {
		if driverOpts[key] != "" {
			return true
		}
	}
	return false
}
