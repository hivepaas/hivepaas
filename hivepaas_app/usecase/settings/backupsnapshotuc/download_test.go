package backupsnapshotuc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup"
)

// A download is a file of the snapshot: a directory is restored, and a name the
// directory does not hold is not found.
func TestDownloadableEntry(t *testing.T) {
	entries := []backup.SnapshotEntry{{Name: "uploads", Dir: true, SizeBytes: 5}, {Name: "db.pg_dump", SizeBytes: 42}}

	entry, err := downloadableEntry(entries, "db.pg_dump")
	assert.NoError(t, err)
	assert.Equal(t, int64(42), entry.SizeBytes)

	_, err = downloadableEntry(entries, "uploads")
	assert.ErrorIs(t, err, hperrors.ErrArgumentInvalid)

	_, err = downloadableEntry(entries, "gone.txt")
	assert.ErrorIs(t, err, hperrors.ErrNotFound)
}

// A download is recorded against the snapshot's app, or its repository's scope
// when no app owns it.
func TestDownloadAuditScope(t *testing.T) {
	repo := &entity.Setting{ID: "r1", Scope: base.ObjectScopeProject, ObjectID: "p1"}

	scope, objectID := downloadAuditScope("a1", repo)
	assert.Equal(t, base.ObjectScopeApp, scope)
	assert.Equal(t, "a1", objectID)

	scope, objectID = downloadAuditScope("", repo)
	assert.Equal(t, base.ObjectScopeProject, scope)
	assert.Equal(t, "p1", objectID)
}
