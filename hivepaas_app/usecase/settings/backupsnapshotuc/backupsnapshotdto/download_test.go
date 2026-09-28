package backupsnapshotdto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A download names a file inside the snapshot, by a relative path.
func TestDownloadRequestNamesAFileInside(t *testing.T) {
	req := NewDownloadBackupSnapshotFileReq()
	req.ID = snapshotRecordID
	req.Path = "./uploads//a.png"
	assert.NoError(t, req.ModifyRequest())
	assert.Empty(t, req.Validate())
	assert.Equal(t, "uploads/a.png", req.Path)

	for _, bad := range []string{"", "/etc/passwd", "../x", "uploads/../../x"} {
		req = NewDownloadBackupSnapshotFileReq()
		req.ID = snapshotRecordID
		req.Path = bad
		assert.NoError(t, req.ModifyRequest())
		assert.NotEmpty(t, req.Validate(), bad)
	}
}
