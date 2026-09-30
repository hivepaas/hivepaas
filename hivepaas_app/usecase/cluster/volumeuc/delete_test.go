package volumeuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
)

type countingFiles struct {
	fileservice.Service
	count int
}

func (c *countingFiles) CountOnVolume(context.Context, database.IDB, string) (int, error) {
	return c.count, nil
}

// A volume still holding HivePaaS's files - caches, job outputs, an app's data
// files - is not deleted from under them.
func TestAVolumeHoldingFilesIsNotDeleted(t *testing.T) {
	err := refuseVolumeWithFiles(context.Background(), nil, &countingFiles{count: 3},
		&entity.Setting{ID: "vol-1", Name: "shop-data"})

	assert.ErrorIs(t, err, hperrors.ErrVolumeHasFiles)
}

func TestAnEmptyVolumeIsDeleted(t *testing.T) {
	assert.NoError(t, refuseVolumeWithFiles(context.Background(), nil, &countingFiles{},
		&entity.Setting{ID: "vol-1"}))
}
