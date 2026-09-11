package auditserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

type fakeAuditLogRepo struct {
	repository.AuditLogRepo
	inserted *entity.AuditLog
}

func (f *fakeAuditLogRepo) Insert(
	_ context.Context, _ database.IDB, auditLog *entity.AuditLog, _ ...bunex.InsertQueryOption,
) error {
	f.inserted = auditLog
	return nil
}

// Section is a column, not a field inside Detail, so it has to be carried across
// rather than left to whatever the caller happened to put in the JSON.
func TestRecordCarriesTheSectionIntoItsOwnColumn(t *testing.T) {
	repo := &fakeAuditLogRepo{}
	svc := New(repo)

	err := svc.Record(context.Background(), nil, &auditservice.Entry{
		Type:    base.AuditLogTypeAppUpdate,
		Scope:   base.ObjectScopeApp,
		Source:  base.AuditLogSourceAPIUpdate,
		Section: "routing",
		Result:  base.AuditLogResultAllowed,
		Detail:  `{"onProbation":true}`,
	})

	assert.NoError(t, err)
	assert.Equal(t, "routing", repo.inserted.Section)
	// The listing drops Detail once it passes a hundred characters, which is why
	// the section had to stop living in there.
	assert.NotContains(t, repo.inserted.Detail, "section")
}

// An entry for a type that covers one endpoint has no section, and an empty
// string must reach the column as NULL rather than as an indexed empty value.
func TestRecordLeavesTheSectionEmptyWhenThereIsNone(t *testing.T) {
	repo := &fakeAuditLogRepo{}
	svc := New(repo)

	err := svc.Record(context.Background(), nil, &auditservice.Entry{
		Type:   base.AuditLogTypeSecretReveal,
		Scope:  base.ObjectScopeGlobal,
		Result: base.AuditLogResultAllowed,
	})

	assert.NoError(t, err)
	assert.Empty(t, repo.inserted.Section)
}
