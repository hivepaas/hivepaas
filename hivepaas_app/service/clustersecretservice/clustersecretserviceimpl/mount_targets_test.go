package clustersecretserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

// A mount at the path goes; an ordinary file there keeps it; the comparison is
// of where the files land, not of how they are named.
func TestMakeRoomComparesResolvedTargets(t *testing.T) {
	spec := &swarm.ContainerSpec{Secrets: []*swarm.SecretReference{
		{SecretID: "mount_key", File: &swarm.SecretReferenceFileTarget{Name: "/run/secrets/app_key"}},
		{SecretID: "db", File: &swarm.SecretReferenceFileTarget{Name: "db_password"}},
	}}
	mounted := map[string]bool{"mount_key": true}

	assert.False(t, makeRoomForSecret(spec, "/run/secrets/app_key", mounted), "the mount steps aside")
	assert.Len(t, spec.Secrets, 1)
	assert.True(t, makeRoomForSecret(spec, "/run/secrets/db_password", mounted),
		"an ordinary secret already there, named relatively")
	assert.False(t, makeRoomForSecret(spec, "/run/secrets/other", mounted))

	cfg := &swarm.ContainerSpec{Configs: []*swarm.ConfigReference{
		{ConfigID: "mount_cert", File: &swarm.ConfigReferenceFileTarget{Name: "/app.pem"}},
	}}
	assert.False(t, makeRoomForConfig(cfg, "/app.pem", map[string]bool{"mount_cert": true}))
	assert.Empty(t, cfg.Configs)
}
