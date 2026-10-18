package appmanageduc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

type fakeApps struct {
	appservice.Service
	app *entity.App
}

func (f *fakeApps) LoadApp(context.Context, database.IDB, string, string, bool, bool,
	...bunex.SelectQueryOption) (*entity.App, error) {
	return f.app, nil
}

// fakeTasks answers each job's last run that ended.
type fakeTasks struct {
	repository.TaskRepo
	lastByJob map[string]*entity.Task
	asked     []string
}

func (f *fakeTasks) ListByTarget(_ context.Context, _ database.IDB, targetID string, _ *basedto.Paging,
	_ ...bunex.SelectQueryOption) ([]*entity.Task, *basedto.PagingMeta, error) {
	f.asked = append(f.asked, targetID)
	if task := f.lastByJob[targetID]; task != nil {
		return []*entity.Task{task}, nil, nil
	}
	return nil, nil, nil
}

type fakeEngine struct {
	managedmodel.Engine
	req *managedmodel.StatusReq
}

func (e *fakeEngine) Status(req *managedmodel.StatusReq) (*managedmodel.Status, error) {
	e.req = req
	checked := time.Date(2026, 10, 18, 11, 58, 0, 0, time.UTC)
	return &managedmodel.Status{
		Archiving:   managedmodel.ArchivingStatus{State: managedmodel.ArchivingOK, CheckedAt: &checked},
		Placement:   "off-site",
		Suggestions: []string{managedmodel.SuggestVerifier},
		LastRestoreTest: &managedmodel.RestoreTestStatus{Passed: true, Duration: 95 * time.Second,
			RestoredTo: "hp-test-1"},
	}, nil
}

type fakeManaged struct {
	managedservice.Service
	engine *fakeEngine
}

func (f *fakeManaged) Engine(name templatemodel.ManagedEngine) (managedmodel.Engine, bool) {
	return f.engine, name == templatemodel.ManagedEnginePostgres
}

func TestStatusIsTheEnginesReadingOfItsJobsLastRuns(t *testing.T) {
	health := &entity.Task{ID: "t-health", Status: base.TaskStatusDone,
		EndedAt: time.Date(2026, 10, 18, 11, 58, 0, 0, time.UTC)}
	health.MustSetOutput(&entity.TaskSchedJobOutputs{Outputs: map[string]string{"FAILED_COUNT": "0"}})
	tasks := &fakeTasks{lastByJob: map[string]*entity.Task{"job-1": health}}
	engine := &fakeEngine{}
	uc := &UC{appService: &fakeApps{app: managedApp()}, taskRepo: tasks, managedService: &fakeManaged{engine: engine}}

	resp, err := uc.GetManagedStatus(context.Background(), nil, &appmanageddto.GetManagedStatusReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", AppID: "app-1"})

	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"job-1", "job-2"}, tasks.asked, "the engine's jobs, not the person's")
	if assert.NotNil(t, engine.req) {
		assert.Equal(t, "postgres", engine.req.Managed.Engine)
		assert.Equal(t, map[string]string{"FAILED_COUNT": "0"}, engine.req.Runs["job-1"].Outputs)
		assert.Nil(t, engine.req.Runs["job-2"], "a job that has not run has no run")
	}
	data := resp.Data
	assert.Equal(t, "postgres", data.Engine)
	assert.Equal(t, "ok", data.Archiving.State)
	assert.Equal(t, "off-site", data.Placement)
	assert.Equal(t, []string{"verifier"}, data.Suggestions)
	assert.Nil(t, data.RestoreWindow)
	assert.Equal(t, int64(95), data.LastRestoreTest.DurationSeconds)
}

func TestStatusOfAnOrdinaryAppIsNotFound(t *testing.T) {
	uc := &UC{appService: &fakeApps{app: &entity.App{ID: "app-1"}}, taskRepo: &fakeTasks{},
		managedService: &fakeManaged{engine: &fakeEngine{}}}

	_, err := uc.GetManagedStatus(context.Background(), nil, &appmanageddto.GetManagedStatusReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", AppID: "app-1"})

	assert.ErrorIs(t, err, hperrors.ErrNotFound)
}
