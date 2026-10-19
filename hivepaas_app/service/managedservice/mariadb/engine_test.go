package mariadb

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
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
	managed := &entity.AppManagedSettings{Engine: "mariadb", EngineAPI: 1}

	for name, call := range map[string]func() error{
		"prepare":   func() error { return engine.PrepareDoc(&managedmodel.PrepareReq{}) },
		"configure": func() error { _, err := engine.Configure(&managedmodel.ConfigureReq{}); return err },
		"status":    func() error { _, err := engine.Status(&managedmodel.StatusReq{}); return err },
		"reconfigure": func() error {
			_, err := engine.Reconfigure(&managedmodel.ReconfigureReq{})
			return err
		},
		"finish restore": func() error {
			_, err := engine.FinishRestore(&managedmodel.FinishRestoreReq{})
			return err
		},
		"retention": func() error { _, err := engine.Retention(&managedmodel.RetentionReq{}); return err },
		"mark":      func() error { _, err := engine.MarkRestorePoint(managed, "before-v2"); return err },
		"mark result": func() error {
			_, err := engine.MarkResult("POSITION=mariadb-bin.000003:389\n")
			return err
		},
		"replica": func() error { _, err := engine.Replica(&managedmodel.ReplicaReq{}); return err },
		"promote": func() error { _, err := engine.Promote(&managedmodel.PromoteReq{}); return err },
	} {
		t.Run(name, func(t *testing.T) {
			assert.ErrorIs(t, call(), hperrors.ErrAppManagedUnsupportedByEngine)
		})
	}
}
