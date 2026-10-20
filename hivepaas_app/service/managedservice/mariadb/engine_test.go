package mariadb

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

func TestTheEngineIsMariaDBs(t *testing.T) {
	engine := New()

	assert.Equal(t, templatemodel.ManagedEngineMariaDB, engine.Name())
	minAPI, maxAPI := engine.APIRange()
	assert.Equal(t, 1, minAPI)
	assert.Equal(t, 1, maxAPI)
}

// What the engine does not do yet, it refuses: an app is never made, nor a
// job run, with half an engine.
func TestWhatTheEngineDoesNotDoYetIsRefused(t *testing.T) {
	_, err := New().Promote(&managedmodel.PromoteReq{})

	assert.ErrorIs(t, err, hperrors.ErrAppManagedUnsupportedByEngine)
}

// MariaDB has replicas, and runs its restore test in a verifier app too.
func TestMariaDBsCapabilities(t *testing.T) {
	assert.Equal(t, managedmodel.Capabilities{Replicas: true, VerifierApp: true}, New().Capabilities())
}
