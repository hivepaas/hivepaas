package syscleanupserviceimpl

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type syncingLogging struct {
	loggingservice.Service
	resp *loggingservice.SyncResp
	err  error
}

func (l *syncingLogging) Sync(context.Context, database.IDB) (*loggingservice.SyncResp, error) {
	return l.resp, l.err
}

type syncingRegistry struct {
	registryservice.Service
	resp *registryservice.SyncResp
	err  error
}

func (r *syncingRegistry) Sync(context.Context, database.IDB) (*registryservice.SyncResp, error) {
	return r.resp, r.err
}

// updateTasks answers a system update running, or none.
type updateTasks struct {
	repository.TaskRepo
	running bool
}

func (u *updateTasks) ListByTarget(context.Context, database.IDB, string, *basedto.Paging,
	...bunex.SelectQueryOption) ([]*entity.Task, *basedto.PagingMeta, error) {
	if u.running {
		return []*entity.Task{{ID: "update"}}, nil, nil
	}
	return nil, nil, nil
}

type schedulingQueue struct {
	queue.TaskQueue
	scheduled []*entity.Task
}

func (q *schedulingQueue) ScheduleTask(_ context.Context, tasks ...*entity.Task) error {
	q.scheduled = append(q.scheduled, tasks...)
	return nil
}

type obiCache struct {
	cacherepository.OBISettingsRepo
	invalidated int
}

func (c *obiCache) Invalidate(context.Context) error {
	c.invalidated++
	return nil
}

func syncData(settings *entity.SystemCleanup) *sysCleanupData {
	req := &syscleanupservice.SysCleanupReq{
		TaskExecData:       &queue.TaskExecData{Task: &entity.Task{ID: "t1"}, LogStore: tasklog.NewLocalStore("t1")},
		SysCleanupSettings: settings,
	}
	req.SetCleanupFlagsDefault()
	return &sysCleanupData{SysCleanupReq: req, TaskOutput: &entity.TaskSystemCleanupOutput{}}
}

type syncFixture struct {
	svc      *service
	logging  *syncingLogging
	registry *syncingRegistry
	queue    *schedulingQueue
	cache    *obiCache
	audit    *recordingAudit
	tasks    *updateTasks
}

func newSyncFixture() *syncFixture {
	f := &syncFixture{
		logging:  &syncingLogging{resp: &loggingservice.SyncResp{}},
		registry: &syncingRegistry{resp: &registryservice.SyncResp{}},
		queue:    &schedulingQueue{}, cache: &obiCache{}, audit: &recordingAudit{}, tasks: &updateTasks{},
	}
	f.svc = &service{loggingService: f.logging, registryService: f.registry, taskQueue: f.queue,
		obiSettings: f.cache, auditService: f.audit, taskRepo: f.tasks}
	return f
}

// What the features did is recorded and logged; their deployments are
// scheduled - and the agents told to read OBI's settings again - only once the
// task's transaction commits; a removal is recorded as a person's is.
func TestSystemAppsSyncSchedulesOnceCommitted(t *testing.T) {
	f := newSyncFixture()
	f.logging.resp = &loggingservice.SyncResp{
		Apps: []*entity.SystemAppSyncOutput{
			{Key: "victoria-logs", Name: "Logging backend", AppID: "a1", Action: entity.SystemAppSyncProvisioned},
			{Key: "vlagent", Name: "Logging collector", AppID: "a2", Action: entity.SystemAppSyncRemoved},
			{Key: "registry", Name: "Registry", AppID: "a4", PreviousAppID: "a3", Action: entity.SystemAppSyncRecreated},
		},
		Tasks: []*entity.Task{{ID: "deploy-a1"}}, OBISettingsChanged: true,
	}
	f.registry.resp = &registryservice.SyncResp{Tasks: []*entity.Task{{ID: "deploy-r"}},
		App: &entity.SystemAppSyncOutput{Key: "registry", Name: "Registry", Action: entity.SystemAppSyncUpdated}}
	data := syncData(&entity.SystemCleanup{})

	err, attention := f.svc.sysSyncSystemApps(context.Background(), savepointDB{}, data)
	assert.NoError(t, err)
	assert.NoError(t, attention)

	assert.Len(t, data.TaskOutput.SystemApps.Apps, 4)
	assert.Empty(t, f.queue.scheduled, "not before the commit")
	assert.Zero(t, f.cache.invalidated)
	data.OnPostTxFunc()
	assert.Len(t, f.queue.scheduled, 2)
	assert.Equal(t, 1, f.cache.invalidated)
	if assert.Len(t, f.audit.entries, 2) {
		assert.Equal(t, base.AuditLogTypeAppDelete, f.audit.entries[0].Type)
		assert.Equal(t, "a2", f.audit.entries[0].ResID)
		assert.Contains(t, f.audit.entries[0].Detail, `"removeStorage":false`)
		assert.Equal(t, "a3", f.audit.entries[1].ResID, "the app removed, not the one made again")
	}
}

// A feature whose sync fails is rolled back alone, with what it made in
// docker; the other still syncs, and the run fails.
func TestSystemAppsSyncRollsBackAFailingFeatureAlone(t *testing.T) {
	f := newSyncFixture()
	cleaned := false
	f.logging.resp = &loggingservice.SyncResp{Tasks: []*entity.Task{{ID: "deploy-a1"}},
		Cleanup: func(context.Context) error { cleaned = true; return nil }}
	f.logging.err = errors.New("the backend would not provision")
	f.registry.resp = &registryservice.SyncResp{
		App: &entity.SystemAppSyncOutput{Key: "registry", Name: "Registry", Action: entity.SystemAppSyncNone}}
	data := syncData(&entity.SystemCleanup{})

	err, attention := f.svc.sysSyncSystemApps(context.Background(), savepointDB{}, data)

	assert.ErrorContains(t, err, "the backend would not provision")
	assert.NoError(t, attention)
	assert.True(t, cleaned)
	apps := data.TaskOutput.SystemApps.Apps
	if assert.Len(t, apps, 2) {
		assert.Equal(t, entity.SystemAppSyncFailed, apps[0].Action)
		assert.Equal(t, "registry", apps[1].Key)
	}
	assert.Nil(t, data.OnPostTxFunc, "nothing of the failed feature is scheduled")
}

// What is left to a person fails the run, for its notification to say so.
func TestSystemAppsSyncFailsTheRunOverWhatItReports(t *testing.T) {
	f := newSyncFixture()
	f.logging.resp = &loggingservice.SyncResp{
		Apps: []*entity.SystemAppSyncOutput{{Key: "vlagent", Name: "Logging collector",
			Action: entity.SystemAppSyncReported, Problem: "its service is scaled to zero"}},
		OBI: &entity.OBISyncOutput{Nodes: []*entity.OBINodeSyncOutput{{NodeID: "n1", Hostname: "one",
			Action: entity.OBISyncReported, Problem: "its agent is not running"}}},
	}
	data := syncData(&entity.SystemCleanup{})

	err, attention := f.svc.sysSyncSystemApps(context.Background(), savepointDB{}, data)

	assert.NoError(t, err)
	assert.ErrorIs(t, attention, errSystemAppsNeedAttention)
	assert.ErrorContains(t, attention, "Logging collector: its service is scaled to zero")
	assert.ErrorContains(t, attention, "OBI on node one: its agent is not running")
}

// Switched off in the settings, or a system update running, nothing syncs.
func TestSystemAppsSyncSkips(t *testing.T) {
	f := newSyncFixture()
	f.logging.err = errors.New("must not be called")
	off := syncData(&entity.SystemCleanup{SystemAppsSync: &entity.SystemAppsSync{Enabled: false}})
	err, attention := f.svc.sysSyncSystemApps(context.Background(), savepointDB{}, off)
	assert.NoError(t, errors.Join(err, attention))
	assert.Nil(t, off.TaskOutput.SystemApps)

	f.tasks.running = true
	updating := syncData(&entity.SystemCleanup{})
	err, attention = f.svc.sysSyncSystemApps(context.Background(), savepointDB{}, updating)
	assert.NoError(t, errors.Join(err, attention))
	assert.Equal(t, "a system update is running", updating.TaskOutput.SystemApps.Skipped)
}

// A run that failed only over what it leaves to a person is not run again; one
// that failed besides is, as before.
func TestAttentionAloneIsNotRetried(t *testing.T) {
	attention := errSystemAppsNeedAttention

	alone := syncData(&entity.SystemCleanup{})
	assert.ErrorIs(t, withAttention(alone, nil, attention), errSystemAppsNeedAttention)
	assert.True(t, alone.TaskNonRetryable)

	besides := syncData(&entity.SystemCleanup{})
	err := withAttention(besides, errors.New("pruning the images failed"), attention)
	assert.ErrorIs(t, err, errSystemAppsNeedAttention)
	assert.ErrorContains(t, err, "pruning the images failed")
	assert.False(t, besides.TaskNonRetryable)

	clean := syncData(&entity.SystemCleanup{})
	assert.NoError(t, withAttention(clean, nil, nil))
	assert.False(t, clean.TaskNonRetryable)
}
