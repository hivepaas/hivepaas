package functionautoscaleserviceimpl

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemeventbusservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
	logs "github.com/hivepaas/hivepaas/services/logging"
)

// settings answers List by the type the query asks for, and keeps writes.
type settings struct {
	repository.SettingRepo
	autoscales  []*entity.Setting
	deployments []*entity.Setting
	jobs        []*entity.Setting
	inserted    []*entity.Setting
	updated     []*entity.Setting
}

func (f *settings) List(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	var rows []*entity.Setting
	sql := bunex.ApplySelect(bun.NewDB(nil, pgdialect.New()).NewSelect().Model(&rows), opts...).String()
	switch {
	case strings.Contains(sql, "'app-autoscale'"):
		return f.autoscales, nil, nil
	case strings.Contains(sql, "'app-deployment'"):
		return f.deployments, nil, nil
	case strings.Contains(sql, "'periodic-job'"):
		return f.jobs, nil, nil
	}
	return nil, nil, nil
}

func (f *settings) Insert(_ context.Context, _ database.IDB, s *entity.Setting, _ ...bunex.InsertQueryOption) error {
	f.inserted = append(f.inserted, s)
	return nil
}

func (f *settings) Update(_ context.Context, _ database.IDB, s *entity.Setting, _ ...bunex.UpdateQueryOption) error {
	f.updated = append(f.updated, s)
	return nil
}

type apps struct {
	repository.AppRepo
	list []*entity.App
}

func (f *apps) ListByIDs(_ context.Context, _ database.IDB, _ string, _ []string,
	_ ...bunex.SelectQueryOption) ([]*entity.App, error) {
	return f.list, nil
}

// fakeSwarm has services by id, and the replicas each update set.
type fakeSwarm struct {
	docker.Manager
	services map[string]*swarm.Service
	scaled   map[string]uint64
}

func (f *fakeSwarm) ServiceInspect(_ context.Context, id string, _ ...docker.ServiceInspectOption) (
	*client.ServiceInspectResult, error) {
	svc := f.services[id]
	if svc == nil {
		return nil, hperrors.NewNotFound("Service")
	}
	return &client.ServiceInspectResult{Service: *svc}, nil
}

func (f *fakeSwarm) ServiceUpdateFunc(_ context.Context, id string, _ *swarm.Service,
	fn func(int, *swarm.Service) (bool, error), _ int, _ time.Duration, _ ...docker.ServiceUpdateOption) error {
	read := *f.services[id]
	spec := *read.Spec.Mode.Replicated
	read.Spec.Mode.Replicated = &spec
	ok, err := fn(0, &read)
	if err != nil || !ok {
		return err
	}
	if f.scaled == nil {
		f.scaled = map[string]uint64{}
	}
	f.scaled[id] = *read.Spec.Mode.Replicated.Replicas
	return nil
}

type loads struct {
	loggingservice.Service
	byApp map[string]*logs.InvocationLoad
	err   error
}

func (f *loads) FunctionLoad(_ context.Context, _ database.IDB, _ []string, _, _ time.Time) (
	map[string]*logs.InvocationLoad, error) {
	return f.byApp, f.err
}

type events struct {
	systemeventbusservice.Service
	published int
}

func (f *events) Publish(context.Context, base.SystemEventType, ...string) error {
	f.published++
	return nil
}

type quiet struct{ logging.Logger }

func (quiet) Warnf(string, ...any) {}

type memStates map[string]state

func (m memStates) load(_ context.Context, id string) state { return m[id] }
func (m memStates) save(_ context.Context, id string, st state) {
	m[id] = st
}

func autoscaleOf(appID string, a *entity.AppAutoscale) *entity.Setting {
	s := &entity.Setting{ID: "as-" + appID, Type: base.SettingTypeAppAutoscale, ObjectID: appID,
		Status: base.SettingStatusActive}
	s.MustSetData(a)
	return s
}

func functionOf(appID string, concurrency int) *entity.Setting {
	s := &entity.Setting{ID: "dep-" + appID, Type: base.SettingTypeAppDeployment, ObjectID: appID}
	s.MustSetData(&entity.AppDeploymentSettings{ActiveMethod: base.DeploymentMethodFunction,
		FunctionSource: &entity.DeploymentFunctionSource{MaxConcurrency: concurrency}})
	return s
}

// serviceOf is a function's service, with the log identity a function has.
func serviceOf(id, appID string, replicas uint64) *swarm.Service {
	return &swarm.Service{ID: id, Spec: swarm.ServiceSpec{
		Mode: swarm.ServiceMode{Replicated: &swarm.ReplicatedService{Replicas: &replicas}},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{Labels: map[string]string{"hivepaas.app.id": appID}},
			LogDriver: &swarm.Driver{Name: "json-file",
				Options: map[string]string{"labels": "hivepaas.app.id"}},
		},
	}}
}

type world struct {
	svc      *service
	settings *settings
	swarm    *fakeSwarm
	loads    *loads
	events   *events
}

func newWorld() *world {
	autoscale := &entity.AppAutoscale{Enabled: true, MinReplicas: 1, MaxReplicas: 5, Target: 70,
		ScaleInDelay: timeutil.Duration(5 * time.Minute)}
	w := &world{
		settings: &settings{
			autoscales: []*entity.Setting{autoscaleOf("F1", autoscale), autoscaleOf("F2", autoscale),
				autoscaleOf("F3", autoscale)},
			deployments: []*entity.Setting{functionOf("F1", 10), functionOf("F2", 10), functionOf("F3", 10)},
		},
		swarm: &fakeSwarm{services: map[string]*swarm.Service{
			"s1": serviceOf("s1", "F1", 1), "s2": serviceOf("s2", "F2", 2), "s3": serviceOf("s3", "F3", 0),
		}},
		loads:  &loads{},
		events: &events{},
	}
	w.svc = &service{states: memStates{}, dockerManager: w.swarm, settingRepo: w.settings,
		appRepo: &apps{list: []*entity.App{
			{ID: "F1", Name: "f1", ServiceID: "s1", Status: base.AppStatusActive},
			{ID: "F2", Name: "f2", ServiceID: "s2", Status: base.AppStatusActive},
			{ID: "F3", Name: "f3", ServiceID: "s3", Status: base.AppStatusActive},
		}},
		loggingService: w.loads, systemEventBus: w.events, logger: quiet{},
		now: func() time.Time { return t0 }}
	return w
}

func run(t *testing.T, w *world) *queue.PeriodicExecData {
	t.Helper()
	data := &queue.PeriodicExecData{Task: &entity.Task{ID: "task"}}
	assert.NoError(t, w.svc.Run(context.Background(), data))
	return data
}

// A turned-away call scales its function at once, saving a task that says
// why; a steady one, and a stopped one, are left alone, and save nothing.
func TestRunScalesWhatItMustAndSavesWhy(t *testing.T) {
	w := newWorld()
	w.loads.byApp = map[string]*logs.InvocationLoad{
		"F1": {BusyMs: 30_000, Calls: 40, Throttled: 4},
		"F2": {BusyMs: 6 * 60_000, Calls: 100},          // 6 in flight at 7 an instance: 1 wanted, low clock starts
		"F3": {BusyMs: 600_000, Calls: 9, Throttled: 9}, // stopped
	}
	data := run(t, w)

	assert.Equal(t, map[string]uint64{"s1": 2}, w.swarm.scaled)
	assert.True(t, data.SaveTask)
	out := &entity.TaskFunctionAutoscaleOutput{}
	assert.NoError(t, json.Unmarshal([]byte(data.Task.Output), out))
	if assert.Len(t, out.Scaled, 1) {
		assert.Equal(t, "F1", out.Scaled[0].App)
		assert.Equal(t, 1, out.Scaled[0].From)
		assert.Equal(t, 2, out.Scaled[0].To)
		assert.Equal(t, int64(4), out.Scaled[0].Throttled)
		assert.Contains(t, out.Scaled[0].Reason, "turned away")
	}

	w.loads.byApp = map[string]*logs.InvocationLoad{"F2": {BusyMs: 13 * 60_000, Calls: 100}}
	w.swarm.scaled = nil
	data = run(t, w)
	assert.Empty(t, w.swarm.scaled, "F2 needs 2: as it is")
	assert.False(t, data.SaveTask, "a run that changed nothing saves no task")
}

// When the calls cannot be read, nothing moves: no data is not no load.
func TestRunHoldsWhenTheLogsCannotBeRead(t *testing.T) {
	w := newWorld()
	w.loads.err = errors.New("logging not enabled")
	data := &queue.PeriodicExecData{Task: &entity.Task{ID: "task"}}
	assert.Error(t, w.svc.Run(context.Background(), data))
	assert.Empty(t, w.swarm.scaled)
	assert.False(t, data.SaveTask)
}

// A function whose logs carry no identity is passed over: its calls would
// read as none.
func TestRunPassesOverAFunctionWithoutItsLogIdentity(t *testing.T) {
	w := newWorld()
	w.swarm.services["s1"].Spec.TaskTemplate.LogDriver = &swarm.Driver{Name: "local"}
	w.loads.byApp = map[string]*logs.InvocationLoad{"F1": {Calls: 10, Throttled: 10}}
	run(t, w)
	assert.Empty(t, w.swarm.scaled)
}

// The job is on while a function has autoscale on, off when none has; made
// the first time, and the workers told each time it moves.
func TestEnsureJob(t *testing.T) {
	w := newWorld()
	assert.NoError(t, w.svc.EnsureJob(context.Background(), nil))
	if assert.Len(t, w.settings.inserted, 1) {
		job := w.settings.inserted[0]
		assert.Equal(t, base.SettingTypePeriodicJob, job.Type)
		assert.Equal(t, string(base.PeriodicKindFunctionAutoscale), job.Kind)
		assert.Equal(t, base.SettingStatusActive, job.Status)
		assert.Equal(t, timeutil.Duration(15*time.Second), job.MustAsPeriodicJob().Interval)
	}
	assert.Equal(t, 1, w.events.published)

	w.settings.jobs = w.settings.inserted
	assert.NoError(t, w.svc.EnsureJob(context.Background(), nil))
	assert.Equal(t, 1, w.events.published, "already on: nothing to tell")

	w.settings.autoscales = nil
	assert.NoError(t, w.svc.EnsureJob(context.Background(), nil))
	if assert.Len(t, w.settings.updated, 1) {
		assert.Equal(t, base.SettingStatusDisabled, w.settings.updated[0].Status)
	}
	assert.Equal(t, 2, w.events.published)
}
