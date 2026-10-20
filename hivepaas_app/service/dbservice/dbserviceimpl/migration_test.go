package dbserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// The table was made with a row named v1-20251019, which sorts after every
// version code: read as the code it stands for, the data it records is
// v000001's, and migration runs.
func TestTheDataMigrationRecordedIsReadAsAVersionCode(t *testing.T) {
	assert.Equal(t, base.VersionCodeV1, appliedCode("v1-20251019"))
	assert.Equal(t, base.VersionCodeV2, appliedCode(base.VersionCodeV2))
	assert.True(t, needsMigration("v1-20251019"))
	assert.True(t, needsMigration(base.VersionCodeV2))
	assert.False(t, needsMigration(base.CurrentVersion))
	assert.False(t, needsMigration("v999999"), "a newer HivePaaS's data is left alone")
}
