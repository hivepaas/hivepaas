package syscleanupserviceimpl

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// savepointDB runs each savepoint's function; a failing one is simply returned,
// as a rolled-back savepoint is.
type savepointDB struct {
	database.IDB
}

func (savepointDB) RunInTx(ctx context.Context, _ *sql.TxOptions, fn func(context.Context, bun.Tx) error) error {
	return fn(ctx, bun.Tx{})
}

// orphanRepo lists the orphans it holds, and gives back each one not yet
// deleted - the way the locked re-read finds a child its parent took with it.
type orphanRepo struct {
	repository.AppRepo
	apps    []*entity.App
	deleted map[string]bool
}

func (r *orphanRepo) List(_ context.Context, _ database.IDB, _ string, _ *basedto.Paging,
	_ ...bunex.SelectQueryOption) ([]*entity.App, *basedto.PagingMeta, error) {
	return r.apps, nil, nil
}

func (r *orphanRepo) GetByID(_ context.Context, _ database.IDB, _, id string,
	_ ...bunex.SelectQueryOption) (*entity.App, error) {
	for _, app := range r.apps {
		if app.ID == id && !r.deleted[id] {
			return app, nil
		}
	}
	return nil, hperrors.NewNotFound("App")
}

type deletingAppService struct {
	appservice.Service
	repo    *orphanRepo
	failing string
	calls   []string
}

func (s *deletingAppService) DeleteApp(_ context.Context, _ database.IDB, app *entity.App,
	removeStorage, cascade bool) error {
	s.calls = append(s.calls, app.ID)
	if removeStorage || !cascade {
		return errors.New("an orphan keeps its data and takes its children")
	}
	if app.ID == s.failing {
		return errors.New("the service would not go")
	}
	s.repo.deleted[app.ID] = true
	for _, child := range s.repo.apps {
		if child.ParentID == app.ID {
			s.repo.deleted[child.ID] = true
		}
	}
	return nil
}

type recordingAudit struct {
	auditservice.Service
	entries []*auditservice.Entry
}

func (a *recordingAudit) Record(_ context.Context, _ database.IDB, entry *auditservice.Entry) error {
	a.entries = append(a.entries, entry)
	return nil
}

// Every orphan is deleted as a person would delete it, and recorded; one that
// fails does not stop the others, and a child its parent took is not tried again.
func TestOrphanedAppsAreDeletedAndRecorded(t *testing.T) {
	repo := &orphanRepo{deleted: map[string]bool{}, apps: []*entity.App{
		{ID: "a1", Name: "api"},
		{ID: "a2", Name: "worker"},
		{ID: "a3", Name: "api-preview", ParentID: "a1"},
	}}
	apps := &deletingAppService{repo: repo, failing: "a2"}
	audit := &recordingAudit{}
	svc := &service{appRepo: repo, appService: apps, auditService: audit}
	data := &sysCleanupData{
		SysCleanupReq: &syscleanupservice.SysCleanupReq{
			TaskExecData: &queue.TaskExecData{LogStore: tasklog.NewLocalStore("test")},
		},
		TaskOutput: &entity.TaskSystemCleanupOutput{DBCleanup: &entity.DBCleanupOutput{}},
	}

	err := svc.sysCleanupDBDeleteOrphanedApps(context.Background(), savepointDB{}, data)

	assert.ErrorContains(t, err, "the service would not go")
	assert.Equal(t, []string{"a1", "a2"}, apps.calls, "the preview went with its parent")
	deleted := data.TaskOutput.DBCleanup.OrphanedAppsDeleted
	if assert.Len(t, deleted, 2) {
		assert.Equal(t, entity.OrphanedAppOutput{ID: "a1", Name: "api"}, *deleted[0])
		assert.Equal(t, "a2", deleted[1].ID)
		assert.NotEmpty(t, deleted[1].Error)
	}
	if assert.Len(t, audit.entries, 2) {
		assert.Equal(t, base.AuditLogTypeAppDelete, audit.entries[0].Type)
		assert.Equal(t, base.AuditLogSourceSystemCleanup, audit.entries[0].Source)
		assert.Contains(t, audit.entries[0].Detail, "orphaned")
	}
}
