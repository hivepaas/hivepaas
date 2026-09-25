package dockerapiservice

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
)

// HivePaaS labels read hivepaas.<object>.<field>, camelCase, as hivepaas.app.id
// does.
func TestDockerAPILabelsFollowTheConvention(t *testing.T) {
	assert.Equal(t, "hivepaas.dockerApi.app", dockerproxy.OwnerLabel)
	assert.Equal(t, "hivepaas.dockerApi.network", NetworkLabel)
	assert.Equal(t, "hivepaas.dockerApi.socket", SocketVolumeLabel)
}
