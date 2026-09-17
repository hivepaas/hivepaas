// Package registry reads what an image registry publishes.
//
// It speaks the OCI distribution API directly rather than through the docker
// daemon, because the daemon can pull an image but cannot say which other tags
// exist. It reads and never writes, and it carries no credentials yet: the
// templates HivePaaS ships are public.
//
// TODO: app templates later - authenticate with the registry-auth setting so a
// private registry can be scanned. See
// docs/superpowers/specs/2026-09-17-app-template-image-override-design.md §10.
package registry

import (
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
)

const (
	dockerHubHost    = "registry-1.docker.io"
	dockerHubLibrary = "library"
	localhostHost    = "localhost"
)

// Reference is a repository as the V2 API addresses it: a host to talk to and a
// path under /v2/.
type Reference struct {
	Host string
	Name string
}

func (r Reference) String() string {
	return r.Host + "/" + r.Name
}

// ParseRepository normalizes a docker image reference into the host and repository
// the registry API uses. `postgres` is `library/postgres` on Docker Hub, while
// `ghcr.io/owner/app` already says where it lives.
//
// The first path segment is a host when it looks like one - it carries a dot or a
// port, or it is localhost. That is docker's own rule, and it is why `minio/minio`
// is a Docker Hub repository and not a host called minio.
func ParseRepository(image string) Reference {
	name := imageref.Parse(image).Repository
	host := dockerHubHost

	if first, rest, found := strings.Cut(name, "/"); found && isHost(first) {
		host, name = first, rest
	}
	if host == dockerHubHost && !strings.Contains(name, "/") {
		name = dockerHubLibrary + "/" + name
	}
	return Reference{Host: host, Name: name}
}

func isHost(segment string) bool {
	return strings.ContainsAny(segment, ".:") || segment == localhostHost
}
