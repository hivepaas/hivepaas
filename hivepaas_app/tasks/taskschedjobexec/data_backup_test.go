package taskschedjobexec

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A data backup step of a sequence tells the steps after it the snapshot it took.
func TestDataBackupStepOutputs(t *testing.T) {
	outputs := dataBackupOutputs(&entity.SchedJobDataBackupResult{SnapshotID: "k1", SizeBytes: 42})

	assert.Equal(t, map[string]string{"SNAPSHOT_ID": "k1", "SNAPSHOT_SIZE_BYTES": "42"}, outputs)
	assert.Nil(t, dataBackupOutputs(nil))
}
