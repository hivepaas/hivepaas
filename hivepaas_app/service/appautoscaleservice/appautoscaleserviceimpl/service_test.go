package appautoscaleserviceimpl

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
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemeventbusservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
	logs "github.com/hivepaas/hivepaas/services/logging"
)

// settings answers List by the type the query asks for, and keeps writes.
type settings struct {
	repository.SettingRepo
	autoscales  []*entity.Setting
	deployments []*entity.Setting
	routings    []*entity.Setting
	jobs        []*entity.Setting
	inserted    []*entity.Setting
	updated     []*entity.Setting
}

func (f *settings) List(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	var rows []*entity.Setting
	sql := bunex.ApplySelect(bun.NewDB(nil, pgdialect.New()).NewSelect().Model(&rows), opts...).String()
	switch {
	case strings.Contains(sql, "'periodic-job'"):
		return f.jobs, nil, nil
	case strings.Contains(sql, "'app-autoscale'"):
		return f.autoscales, nil, nil
	case strings.Contains(sql, "'app-deployment'"):
		return f.deployments, nil, nil
	case strings.Contains(sql, "'app-routing'"):
		return f.routings, nil, nil
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

// ServiceList answers every service, whatever the filters: the run maps them
// by id.
func (f *fakeSwarm) ServiceList(_ context.Context, _ ...docker.ServiceListOption) (*client.ServiceListResult, error) {
	out := &client.ServiceListResult{}
	for _, svc := range f.services {
		out.Items = append(out.Items, *svc)
	}
	return out, nil
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
	byApp    map[string]*logs.InvocationLoad
	err      error
	requests map[string]*logs.RequestLoad
	cpu      map[string][]*logs.ContainerCPU
	appsErr  error
	// ranges are the ranges read, by what was read.
	ranges map[string][2]time.Time
}

func (f *loads) read(what string, start, end time.Time) {
	if f.ranges == nil {
		f.ranges = map[string][2]time.Time{}
	}
	f.ranges[what] = [2]time.Time{start, end}
}

func (f *loads) FunctionLoad(_ context.Context, _ database.IDB, _ []string, start, end time.Time) (
	map[string]*logs.InvocationLoad, error) {
	f.read("calls", start, end)
	return f.byApp, f.err
}

func (f *loads) RequestLoad(_ context.Context, _ database.IDB, _ []string, start, end time.Time) (
	map[string]*logs.RequestLoad, error) {
	f.read("requests", start, end)
	return f.requests, f.appsErr
}

func (f *loads) CPULoad(_ context.Context, _ database.IDB, _ []string, start, end time.Time) (
	map[string][]*logs.ContainerCPU, error) {
	f.read("cpu", start, end)
	return f.cpu, f.appsErr
}

// proxy is Traefik's service, writing its access log as JSON, labeled.
type proxy struct {
	traefikservice.Service
	args []string
}

func (f *proxy) GetTraefikSwarmService(context.Context) (*swarm.Service, error) {
	return &swarm.Service{Spec: swarm.ServiceSpec{
		Annotations: swarm.Annotations{},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{Args: f.args,
				Labels: map[string]string{base.LabelLogComponent: base.LogComponentTraefik}},
			LogDriver: &swarm.Driver{Name: "json-file", Options: map[string]string{"labels": base.LabelLogComponent}},
		},
	}}, nil
}

// agent is the HivePaaS agent's service, labeled.
type agent struct {
	hpappservice.Service
}

func (agent) GetHpAgentSwarmService(context.Context) (*swarm.Service, error) {
	return &swarm.Service{Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{Labels: map[string]string{base.LabelLogComponent: base.LogComponentAgent}},
		LogDriver:     &swarm.Driver{Name: "json-file", Options: map[string]string{"labels": base.LabelLogComponent}},
	}}}, nil
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
		traefikService: &proxy{args: []string{"--accesslog=true", "--accesslog.format=json"}},
		hpAppService:   agent{},
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
	out := &entity.TaskAppAutoscaleOutput{}
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
		assert.Equal(t, string(base.PeriodicKindAppAutoscale), job.Kind)
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

// appWorld is a world with one app that is not a function: A1, exposed, its
// service at 2 replicas with a CPU limit of a core, running both.
func appWorld(autoscale *entity.AppAutoscale) *world {
	w := newWorld()
	autoscale.Enabled, autoscale.MinReplicas, autoscale.MaxReplicas = true, 1, 10
	autoscale.ScaleInDelay = timeutil.Duration(5 * time.Minute)
	w.settings.autoscales = []*entity.Setting{autoscaleOf("A1", autoscale)}
	w.settings.deployments = nil
	routing := &entity.Setting{ID: "rt-A1", Type: base.SettingTypeAppRouting, ObjectID: "A1"}
	routing.MustSetData(&entity.AppRoutingSettings{ExposePublicly: true,
		Domains: []*entity.AppDomain{{Enabled: true, Domain: "a1.example.com"}}})
	w.settings.routings = []*entity.Setting{routing}
	svc := serviceOf("s1", "A1", 2)
	svc.Spec.TaskTemplate.Resources = &swarm.ResourceRequirements{Limits: &swarm.Limit{NanoCPUs: 1e9}}
	svc.ServiceStatus = &swarm.ServiceStatus{RunningTasks: 2, DesiredTasks: 2}
	w.swarm.services = map[string]*swarm.Service{"s1": svc}
	w.svc.appRepo = &apps{list: []*entity.App{{ID: "A1", Name: "a1", ServiceID: "s1", Status: base.AppStatusActive}}}
	return w
}

func runAt(t *testing.T, w *world, at time.Time) *queue.PeriodicExecData {
	t.Helper()
	w.svc.now = func() time.Time { return at }
	return run(t, w)
}

func onCPU(cores ...float64) []*logs.ContainerCPU {
	out := make([]*logs.ContainerCPU, 0, len(cores))
	for i, c := range cores {
		out = append(out, &logs.ContainerCPU{Container: string(rune('a' + i)), CPU: c, Limit: 1})
	}
	return out
}

// An app on its requests scales to what they need, on the second run above,
// and the task says so.
func TestRunScalesAnAppOnItsRequests(t *testing.T) {
	w := appWorld(&entity.AppAutoscale{RequestsTarget: 10})
	w.loads.requests = map[string]*logs.RequestLoad{"A1": {BusyMs: 35 * 60_000, Requests: 7000}}

	data := runAt(t, w, t0)
	assert.Empty(t, w.swarm.scaled, "one run above is not a trend")
	assert.False(t, data.SaveTask)

	data = runAt(t, w, t0.Add(15*time.Second))
	assert.Equal(t, map[string]uint64{"s1": 4}, w.swarm.scaled, "35 in flight at 10 an instance")
	out := &entity.TaskAppAutoscaleOutput{}
	assert.NoError(t, json.Unmarshal([]byte(data.Task.Output), out))
	if assert.Len(t, out.Scaled, 1) {
		assert.Equal(t, int64(7000), out.Scaled[0].Requests)
		assert.Equal(t, 10, out.Scaled[0].RequestsTarget)
		assert.InDelta(t, 35, out.Scaled[0].InFlight, 1e-9)
		assert.Contains(t, out.Scaled[0].Reason, "35.0 requests in flight, 10 an instance")
	}
}

// An app on its CPU scales by how far it is from its target; then waits a
// minute before scaling out again on CPU.
func TestRunScalesAnAppOnItsCPUWithACooldown(t *testing.T) {
	w := appWorld(&entity.AppAutoscale{CPUTarget: 60})
	w.loads.cpu = map[string][]*logs.ContainerCPU{"A1": onCPU(0.9, 0.9)}

	runAt(t, w, t0)
	data := runAt(t, w, t0.Add(15*time.Second))
	assert.Equal(t, map[string]uint64{"s1": 3}, w.swarm.scaled, "90 % of a core, 60 % wanted: 2 * 1.5")
	out := &entity.TaskAppAutoscaleOutput{}
	assert.NoError(t, json.Unmarshal([]byte(data.Task.Output), out))
	if assert.Len(t, out.Scaled, 1) {
		assert.InDelta(t, 90, out.Scaled[0].CPU, 1e-9)
		assert.Equal(t, 60, out.Scaled[0].CPUTarget)
	}

	// Scaled: 3 replicas, still hot.
	*w.swarm.services["s1"].Spec.Mode.Replicated.Replicas = 3
	w.swarm.services["s1"].ServiceStatus = &swarm.ServiceStatus{RunningTasks: 3, DesiredTasks: 3}
	w.swarm.scaled = nil
	runAt(t, w, t0.Add(30*time.Second))
	runAt(t, w, t0.Add(45*time.Second))
	assert.Empty(t, w.swarm.scaled, "within the cooldown")
	runAt(t, w, t0.Add(75*time.Second))
	assert.Equal(t, map[string]uint64{"s1": 5}, w.swarm.scaled, "the cooldown over")
}

// CPU within a tenth of its target changes nothing.
func TestRunLeavesCPUNearItsTarget(t *testing.T) {
	w := appWorld(&entity.AppAutoscale{CPUTarget: 70})
	w.loads.cpu = map[string][]*logs.ContainerCPU{"A1": onCPU(0.75, 0.75)}
	runAt(t, w, t0)
	runAt(t, w, t0.Add(15*time.Second))
	assert.Empty(t, w.swarm.scaled)
}

// What cannot be read moves nothing: no row from the agent for a running app,
// a query that failed, the access log not JSON, an app no domain reaches.
func TestRunHoldsAnAppWhoseSignalsCannotBeRead(t *testing.T) {
	cases := map[string]func(w *world){
		"no agent row": func(w *world) { w.loads.cpu = nil },
		"query failed": func(w *world) { w.loads.appsErr = errors.New("logging not enabled") },
		"access log not json": func(w *world) {
			w.svc.traefikService = &proxy{args: []string{"--accesslog=true"}}
			w.loads.cpu = nil
		},
		"not exposed": func(w *world) {
			w.settings.routings = nil
			w.loads.cpu = nil
		},
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			w := appWorld(&entity.AppAutoscale{RequestsTarget: 1, CPUTarget: 50})
			w.loads.requests = map[string]*logs.RequestLoad{"A1": {BusyMs: 99 * 60_000}}
			w.loads.cpu = map[string][]*logs.ContainerCPU{"A1": onCPU(0.1, 0.1)}
			setup(w)
			for i := range 30 {
				w.svc.now = func() time.Time { return t0.Add(time.Duration(i) * 15 * time.Second) }
				data := &queue.PeriodicExecData{Task: &entity.Task{ID: "task"}}
				_ = w.svc.Run(context.Background(), data)
			}
			if name == "no agent row" {
				// Requests still read: 99 in flight - but CPU, which
				// cannot be read, does not stop them.
				assert.Equal(t, map[string]uint64{"s1": 10}, w.swarm.scaled)
				return
			}
			assert.Empty(t, w.swarm.scaled)
		})
	}
}

// A service that cannot place what it was scaled to for 2 minutes is not
// scaled further out; it is scaled in as ever.
func TestRunHoldsAScaleOutTheClusterHasNoRoomFor(t *testing.T) {
	w := appWorld(&entity.AppAutoscale{RequestsTarget: 10})
	w.swarm.services["s1"].ServiceStatus = &swarm.ServiceStatus{RunningTasks: 1, DesiredTasks: 2}
	w.loads.requests = map[string]*logs.RequestLoad{"A1": {BusyMs: 50 * 60_000}}
	for i := range 12 {
		runAt(t, w, t0.Add(time.Duration(i)*15*time.Second))
	}
	assert.Equal(t, map[string]uint64{"s1": 5}, w.swarm.scaled, "scaled out at first")

	w.swarm.scaled = nil
	for i := 12; i < 20; i++ {
		runAt(t, w, t0.Add(time.Duration(i)*15*time.Second))
	}
	assert.Empty(t, w.swarm.scaled, "short of tasks for 2 minutes: no further")
}

// An app publishing a port on its node is not scaled.
func TestRunPassesOverAnAppWithHostPorts(t *testing.T) {
	w := appWorld(&entity.AppAutoscale{RequestsTarget: 1})
	w.swarm.services["s1"].Spec.EndpointSpec = &swarm.EndpointSpec{Ports: []swarm.PortConfig{
		{TargetPort: 80, PublishedPort: 80, PublishMode: swarm.PortConfigPublishModeHost}}}
	w.loads.requests = map[string]*logs.RequestLoad{"A1": {BusyMs: 50 * 60_000}}
	runAt(t, w, t0)
	runAt(t, w, t0.Add(15*time.Second))
	assert.Empty(t, w.swarm.scaled)
}

// The job made when it scaled functions only is renamed, not made again.
func TestEnsureJobRenamesTheFunctionsJob(t *testing.T) {
	w := newWorld()
	old := &entity.Setting{ID: "job", Type: base.SettingTypePeriodicJob, Kind: legacyJobKind,
		Name: "Function autoscale", Status: base.SettingStatusActive}
	w.settings.jobs = []*entity.Setting{old}
	assert.NoError(t, w.svc.EnsureJob(context.Background(), nil))
	assert.Empty(t, w.settings.inserted)
	if assert.Len(t, w.settings.updated, 1) {
		assert.Equal(t, string(base.PeriodicKindAppAutoscale), w.settings.updated[0].Kind)
		assert.Equal(t, jobName, w.settings.updated[0].Name)
	}
	assert.Equal(t, 1, w.events.published)
}

// Every signal is read over the same minute, ending 10 s before now: the
// lines of the last seconds may not have arrived from every node.
func TestRunReadsAMinuteEndingBeforeNow(t *testing.T) {
	want := [2]time.Time{t0.Add(-70 * time.Second), t0.Add(-10 * time.Second)}

	w := newWorld()
	run(t, w)
	assert.Equal(t, want, w.loads.ranges["calls"])

	w = appWorld(&entity.AppAutoscale{RequestsTarget: 10, CPUTarget: 70})
	runAt(t, w, t0)
	assert.Equal(t, want, w.loads.ranges["requests"])
	assert.Equal(t, want, w.loads.ranges["cpu"])
}
