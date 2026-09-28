package backupsnapshotuc

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup"
)

// A snapshot the repository no longer holds counts as deleted; any other
// failure stands, and the record stays for another try.
func TestSnapshotGoneIsDeleted(t *testing.T) {
	gone := hperrors.Wrap(fmt.Errorf("%w: k1", backup.ErrSnapshotNotFound))
	assert.NoError(t, snapshotGoneIsDeleted(gone))
	assert.NoError(t, snapshotGoneIsDeleted(nil))

	refused := hperrors.Wrap(fmt.Errorf("%w: invalid repository password", backup.ErrCommandFailed))
	assert.True(t, errors.Is(snapshotGoneIsDeleted(refused), backup.ErrCommandFailed))
}
