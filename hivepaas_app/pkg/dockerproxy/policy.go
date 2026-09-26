// Package dockerproxy gives an app the Docker API without the Docker socket.
//
// An app that starts containers of its own - a CI runner, Autobase, Appwrite's
// executor - is installed upstream by mounting /var/run/docker.sock into it,
// which makes it root on its node. The proxy stands between such an app and the
// daemon: it lets through what the app's policy allows, rewrites what it must,
// and refuses the rest with a Docker error the app logs like any other.
//
// Everything is refused unless a rule lets it through. See
// docs/superpowers/specs/2026-09-24-docker-api-access-design.md.
package dockerproxy

import (
	"slices"
	"strings"
)

// Group names endpoints beyond the core that a policy may allow.
type Group string

const (
	// GroupExec is running commands in the app's children.
	GroupExec Group = "exec"
	// GroupFiles is copying files into and out of the app's children.
	GroupFiles Group = "files"
	// GroupVolumes is volumes of the app's own, and naming them in mounts.
	GroupVolumes Group = "volumes"
	// GroupNetworks is networks of the app's own, and connecting to them.
	GroupNetworks Group = "networks"
	// GroupNestedSocket lets a child mount the app's socket, so that what the
	// child starts goes through this proxy under the same policy.
	GroupNestedSocket Group = "nestedSocket"
)

const (
	// OwnerLabel marks everything the proxy creates for an app, with the app's id.
	OwnerLabel = "hivepaas.dockerApi.app"
	// SocketPath is where an app with access finds its socket.
	SocketPath = "/var/run/hivepaas/docker.sock"
	// SocketFile is the socket's name inside the app's socket volume.
	SocketFile = "docker.sock"
)

// Limits bound what an app's children may use.
type Limits struct {
	// Containers is how many children may exist at once, running or not.
	Containers int
	// Memory is the most one child may ask for, in bytes, and what it gets when
	// it asks for nothing.
	Memory int64
	// NanoCPUs is the same for processor time, in billionths of a CPU.
	NanoCPUs int64
}

// Policy is what one app may do through its socket.
type Policy struct {
	// AppID is written into OwnerLabel on everything created for the app.
	AppID string
	// ServiceID is the app's swarm service. Its task containers are the app
	// itself: they hold the mounts shared directories are found in, and they may
	// connect themselves to the app's networks.
	ServiceID string
	// Images are patterns over the images children may run (see matchImage).
	Images []string
	// SharedDirs are directories of the app a child may bind. Each lies on one of
	// the app's volume mounts.
	SharedDirs []string
	// SharedVolumes are volume names that stand for one of SharedDirs each, for
	// an app whose code names a volume rather than a path: Appwrite's
	// orchestrator mounts "appwrite-builds" into every build. A child naming one
	// gets the directory it stands for, which is what a bind of the path would
	// get - never a volume of that name on the node.
	SharedVolumes map[string]string
	// Network is the app's own network, which children join unless they name
	// another the policy allows.
	Network string
	// Networks are other networks children may join, by name.
	Networks []string
	// SocketVolume is the volume holding the app's socket on every node.
	SocketVolume string
	// ReservedPrefix names the volumes and networks HivePaaS makes for apps -
	// each app's socket volume and its own network. A child may not create a
	// name that starts with it: the name is what says whose socket that is, and
	// whose children a network holds, so one an app could take would be a way to
	// be handed another app's, or to hand its own away. Its own socket volume is
	// reached through nestedSocket, not by naming it.
	ReservedPrefix string
	// Allow are the groups of endpoints beyond the core.
	Allow  []Group
	Limits Limits
}

func (p *Policy) allows(group Group) bool {
	return group == "" || slices.Contains(p.Allow, group)
}

// sharedVolume is the directory a volume name stands for, and whether it stands
// for one.
func (p *Policy) sharedVolume(name string) (string, bool) {
	dir, found := p.SharedVolumes[name]
	return dir, found
}

// reserved reports a name HivePaaS keeps for itself.
func (p *Policy) reserved(name string) bool {
	return p.ReservedPrefix != "" && strings.HasPrefix(name, p.ReservedPrefix)
}

// joinable reports whether a child may join the network of that name.
func (p *Policy) joinable(name string) bool {
	return name == p.Network || slices.Contains(p.Networks, name)
}
