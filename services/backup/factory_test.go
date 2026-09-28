package backup

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

// An engine reaches a repository on cloud storage, on a volume, or through a
// repository server.
func TestNewEngineTakesEveryStorage(t *testing.T) {
	storages := map[string]*Storage{
		"s3":     {StorageS3: &StorageS3{Bucket: "b"}},
		"local":  {StorageLocal: &StorageLocal{Path: "/srv/r"}},
		"server": {StorageServer: &StorageServer{URL: "https://10.0.1.5:40123"}},
	}
	for name, storage := range storages {
		_, err := NewEngine(EngineTypeKopia, storage, DefaultCommandExecutor)
		assert.NoError(t, err, name)
	}

	_, err := NewEngine(EngineTypeKopia, &Storage{}, DefaultCommandExecutor)
	assert.ErrorIs(t, err, backupmodel.ErrStorageTypeRequired)
}
