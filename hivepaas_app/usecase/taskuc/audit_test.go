package taskuc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

// spyAuditService keeps what was recorded, so a test can assert on the record
// itself rather than only on the answer the caller got.
type spyAuditService struct {
	entries []*auditservice.Entry
	err     error
}

func (s *spyAuditService) Record(_ context.Context, _ database.IDB, entry *auditservice.Entry) error {
	s.entries = append(s.entries, entry)
	return s.err
}

func newUCTest() (*UC, *spyAuditService) {
	audit := &spyAuditService{}
	return &UC{auditService: audit}, audit
}

func adminAuth() *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{
		User: &entity.User{ID: "admin-user", Username: "admin"},
	}}
}

func appTask() *entity.Task {
	return &entity.Task{
		ID:       "task-1",
		Scope:    base.ObjectScopeApp,
		ObjectID: "app-1",
		Type:     base.TaskTypeAppDeploy,
		Status:   base.TaskStatusInProgress,
	}
}

// The task's own scope, not the caller's: a project's caller may cancel an app's
// work, and it is the app whose timeline has to show it.
func TestRecordTaskCancelFilesUnderTheTasksOwnObject(t *testing.T) {
	tests := []struct {
		name string
		task *entity.Task
	}{
		{name: "an app's work", task: appTask()},
		{
			name: "a project's work",
			task: &entity.Task{
				ID: "task-2", Scope: base.ObjectScopeProject, ObjectID: "proj-1",
				Type: base.TaskTypeAppDeploy,
			},
		},
		{
			// No object to file under, and none needed: a cleanup sweep or a
			// system backup belongs to the install rather than to a row. The
			// global scope is the empty scope type.
			name: "work under no object",
			task: &entity.Task{ID: "task-3", Scope: base.ObjectScopeGlobal, Type: base.TaskTypeSystemUpdate},
		},
		{
			name: "the install's own work",
			task: &entity.Task{ID: "task-4", Scope: base.ObjectScopeHivepaas, Type: base.TaskTypeSettingsRevert},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, audit := newUCTest()

			err := uc.recordTaskCancel(context.Background(), nil, adminAuth(), tt.task, true)

			assert.NoError(t, err)
			assert.Len(t, audit.entries, 1)
			entry := audit.entries[0]
			assert.Equal(t, base.AuditLogTypeTaskCancel, entry.Type)
			assert.Equal(t, tt.task.Scope, entry.Scope)
			assert.Equal(t, tt.task.ObjectID, entry.ObjectID)
			assert.Equal(t, base.AuditLogSourceAPIAction, entry.Source)
			assert.Equal(t, base.AuditLogResultAllowed, entry.Result)
			assert.Equal(t, base.ResourceTypeTask, entry.ResType)
			assert.Equal(t, tt.task.ID, entry.ResID)
			// What was stopped, not just that something was.
			assert.Equal(t, string(tt.task.Type), entry.ResName)
			assert.Contains(t, entry.Detail, string(tt.task.Type))
			assert.Contains(t, entry.Detail, `"canceled":true`)
		})
	}
}

// A task already running is only sent the request, and the entry has to say so:
// read as a finished cancel it would claim the work stopped when it may not have.
func TestRecordTaskCancelSaysWhenTheWorkWasOnlyAskedToStop(t *testing.T) {
	uc, audit := newUCTest()

	err := uc.recordTaskCancel(context.Background(), nil, adminAuth(), appTask(), false)

	assert.NoError(t, err)
	assert.Len(t, audit.entries, 1)
	assert.Contains(t, audit.entries[0].Detail, `"canceled":false`)
}

func TestRecordTaskCancelRefusesAnEntryThatAnswersNothing(t *testing.T) {
	tests := []struct {
		name string
		auth *basedto.Auth
		task *entity.Task
	}{
		// An entry naming nobody answers none of the questions the entry exists for.
		{name: "no actor", auth: nil, task: appTask()},
		{name: "no task", auth: adminAuth(), task: nil},
		{name: "a task with no id", auth: adminAuth(), task: &entity.Task{Scope: base.ObjectScopeApp}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, audit := newUCTest()

			err := uc.recordTaskCancel(context.Background(), nil, tt.auth, tt.task, true)

			assert.Error(t, err)
			assert.Empty(t, audit.entries)
		})
	}
}

// The caller applies the cancel only if this returns nil, so a store that is down
// has to come back as an error rather than as a silent success.
func TestRecordTaskCancelFailsWhenTheStoreIsDown(t *testing.T) {
	uc, audit := newUCTest()
	audit.err = errors.New("audit store is down")

	err := uc.recordTaskCancel(context.Background(), nil, adminAuth(), appTask(), true)

	assert.Error(t, err)
}
