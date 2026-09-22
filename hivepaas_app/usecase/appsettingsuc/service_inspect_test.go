package appsettingsuc

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// An app whose swarm service is gone - never deployed, or removed by hand - used
// to reach the settings screens as a nil service, which they dereferenced. Each
// of the four screens reads the service, so each of them crashed the request.
func TestRequireServiceRefusesAMissingOne(t *testing.T) {
	_, err := requireAppService(nil, "app-1")

	assert.Error(t, err)
	assert.ErrorIs(t, err, hperrors.ErrAppServiceUnavailable)
}

func TestRequireServicePassesOneThrough(t *testing.T) {
	service := &swarm.Service{ID: "svc-1"}

	got, err := requireAppService(service, "app-1")

	assert.NoError(t, err)
	assert.Same(t, service, got)
}
