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
	engine := New()

	for name, call := range map[string]func() error{
		"replica": func() error { _, err := engine.Replica(&managedmodel.ReplicaReq{}); return err },
		"promote": func() error { _, err := engine.Promote(&managedmodel.PromoteReq{}); return err },
	} {
		t.Run(name, func(t *testing.T) {
			assert.ErrorIs(t, call(), hperrors.ErrAppManagedUnsupportedByEngine)
		})
	}
}
