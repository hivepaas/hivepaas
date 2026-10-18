package taskschedjobexec

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
)

func TestTheArtifactsRunNamesTheAppsThatFailed(t *testing.T) {
	log := reconcileLog(&artifactservice.ReconcileAllResp{Apps: 2, Failed: []*artifactservice.ReconcileFailure{
		{App: &entity.App{Key: "db3"}, Err: hperrors.Wrap(hperrors.ErrArtifactArchUnsupported).
			WithParam("Name", "wal-g").WithParam("Node", "node-2").WithParam("Arch", "riscv64")},
	}})

	assert.Contains(t, log, "The artifacts of 2 app(s) are on every node they may start on.\n")
	assert.Contains(t, log, "db3: ")
	assert.Contains(t, log, "riscv64")
}
