package appautoscaleserviceimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
	logs "github.com/hivepaas/hivepaas/services/logging"
)

// calls counts what a run asks of the store, Docker, the logs and Redis - the
// round trips. A load is a query of VictoriaLogs, and of the logging setting
// before it.
type calls struct {
	db, docker, loads, redis, scaled int
}

var benchDB = bun.NewDB(nil, pgdialect.New())

type countedSettings struct {
	*settings
	n *calls
}

func (f countedSettings) List(ctx context.Context, db database.IDB, scope *entity.ObjectScope, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	f.n.db++
	// The query is built as the store builds it, on one database: the fake's
	// List makes one a call, which costs more than the run.
	var rows []*entity.Setting
	sql := bunex.ApplySelect(benchDB.NewSelect().Model(&rows), opts...).String()
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

func (f countedSettings) Update(ctx context.Context, db database.IDB, s *entity.Setting,
	opts ...bunex.UpdateQueryOption) error {
	f.n.db++
	return f.settings.Update(ctx, db, s, opts...)
}

type countedApps struct {
	*apps
	n *calls
}

func (f countedApps) ListByIDs(ctx context.Context, db database.IDB, projectID string, ids []string,
	opts ...bunex.SelectQueryOption) ([]*entity.App, error) {
	f.n.db++
	return f.apps.ListByIDs(ctx, db, projectID, ids, opts...)
}

type countedSwarm struct {
	*fakeSwarm
	n *calls
}

func (f countedSwarm) ServiceList(ctx context.Context, opts ...docker.ServiceListOption) (
	*client.ServiceListResult, error) {
	f.n.docker++
	return f.fakeSwarm.ServiceList(ctx, opts...)
}

func (f countedSwarm) ServiceUpdateFunc(ctx context.Context, id string, svc *swarm.Service,
	fn func(int, *swarm.Service) (bool, error), retries int, delay time.Duration,
	opts ...docker.ServiceUpdateOption) error {
	f.n.docker++
	f.n.scaled++
	return f.fakeSwarm.ServiceUpdateFunc(ctx, id, svc, fn, retries, delay, opts...)
}

type countedProxy struct {
	*proxy
	n *calls
}

func (f countedProxy) GetTraefikSwarmService(ctx context.Context) (*swarm.Service, error) {
	f.n.docker++
	return f.proxy.GetTraefikSwarmService(ctx)
}

type countedAgent struct {
	agent
	n *calls
}

func (f countedAgent) GetHpAgentSwarmService(ctx context.Context) (*swarm.Service, error) {
	f.n.docker++
	return f.agent.GetHpAgentSwarmService(ctx)
}

type countedLoads struct {
	*loads
	n *calls
}

func (f countedLoads) FunctionLoad(ctx context.Context, db database.IDB, ids []string,
	start, end, shortStart time.Time) (map[string]*logs.InvocationLoad, error) {
	f.n.loads++
	return f.loads.FunctionLoad(ctx, db, ids, start, end, shortStart)
}

func (f countedLoads) RequestLoad(ctx context.Context, db database.IDB, ids []string,
	start, end, shortStart time.Time) (map[string]*logs.RequestLoad, error) {
	f.n.loads++
	return f.loads.RequestLoad(ctx, db, ids, start, end, shortStart)
}

func (f countedLoads) CPULoad(ctx context.Context, db database.IDB, ids []string,
	start, end time.Time) (map[string][]*logs.ContainerCPU, error) {
	f.n.loads++
	return f.loads.CPULoad(ctx, db, ids, start, end)
}

// jsonStates keeps the states as Redis does: marshaled, by key. A call is a
// round trip.
type jsonStates struct {
	raw map[string][]byte
	n   *calls
}

func (m jsonStates) load(_ context.Context, ids []string) map[string]state {
	m.n.redis++
	out := map[string]state{}
	for _, id := range ids {
		if raw, found := m.raw[fmt.Sprintf(stateKeyFmt, id)]; found {
			var st state
			if json.Unmarshal(raw, &st) == nil {
				out[id] = st
			}
		}
	}
	return out
}

func (m jsonStates) save(_ context.Context, states map[string]state) {
	if len(states) == 0 {
		return
	}
	m.n.redis++
	for id, st := range states {
		key := fmt.Sprintf(stateKeyFmt, id)
		if st == (state{}) {
			delete(m.raw, key)
			continue
		}
		m.raw[key], _ = json.Marshal(st)
	}
}

// benchWorld is n apps with autoscale on - a third functions, a third apps on
// their requests, a third on their CPU - each at 2 replicas, its load near
// its target: a run scales none, as most runs do.
func benchWorld(n int) (*world, *calls) {
	w := newWorld()
	counted := &calls{}
	w.settings.autoscales, w.settings.deployments, w.settings.routings = nil, nil, nil
	w.settings.jobs = []*entity.Setting{{ID: "job", Type: base.SettingTypePeriodicJob,
		Kind: string(base.PeriodicKindAppAutoscale), Status: base.SettingStatusActive}}
	w.swarm.services = map[string]*swarm.Service{}
	w.loads.byApp = map[string]*logs.InvocationLoad{}
	w.loads.requests = map[string]*logs.RequestLoad{}
	w.loads.cpu = map[string][]*logs.ContainerCPU{}
	list := make([]*entity.App, 0, n)
	for i := range n {
		id, svcID := fmt.Sprintf("A%d", i), fmt.Sprintf("s%d", i)
		autoscale := &entity.AppAutoscale{Enabled: true, MinReplicas: 1, MaxReplicas: 10, Target: 70}
		svc := serviceOf(svcID, id, 2)
		svc.ServiceStatus = &swarm.ServiceStatus{RunningTasks: 2, DesiredTasks: 2}
		switch i % 3 {
		case 0:
			w.settings.deployments = append(w.settings.deployments, functionOf(id, 8))
			w.loads.byApp[id] = &logs.InvocationLoad{BusyMs: 11 * 60_000, Calls: 600}
		case 1:
			autoscale.RequestsTarget = 10
			routing := &entity.Setting{ID: "rt-" + id, Type: base.SettingTypeAppRouting, ObjectID: id}
			routing.MustSetData(&entity.AppRoutingSettings{ExposePublicly: true,
				Domains: []*entity.AppDomain{{Enabled: true, Domain: id + ".example.com"}}})
			w.settings.routings = append(w.settings.routings, routing)
			w.loads.requests[id] = &logs.RequestLoad{BusyMs: 18 * 60_000, Requests: 3000}
		default:
			autoscale.CPUTarget = 70
			svc.Spec.TaskTemplate.Resources = &swarm.ResourceRequirements{Limits: &swarm.Limit{NanoCPUs: 1e9}}
			w.loads.cpu[id] = onCPU(0.7, 0.72)
		}
		w.settings.autoscales = append(w.settings.autoscales, autoscaleOf(id, autoscale))
		w.swarm.services[svcID] = svc
		list = append(list, &entity.App{ID: id, Name: id, ServiceID: svcID, Status: base.AppStatusActive})
	}
	w.svc.settingRepo = countedSettings{settings: w.settings, n: counted}
	w.svc.appRepo = countedApps{apps: &apps{list: list}, n: counted}
	w.svc.dockerManager = countedSwarm{fakeSwarm: w.swarm, n: counted}
	w.svc.traefikService = countedProxy{proxy: &proxy{args: []string{"--accesslog=true", "--accesslog.format=json"}},
		n: counted}
	w.svc.hpAppService = countedAgent{n: counted}
	w.svc.loggingService = countedLoads{loads: w.loads, n: counted}
	w.svc.states = jsonStates{raw: map[string][]byte{}, n: counted}
	return w, counted
}

// BenchmarkRun is what a run costs this process with n apps on autoscale, and
// what it asks of the others, a run each. With none, it is the one run that
// turns the job off: no other runs then.
func BenchmarkRun(b *testing.B) {
	for _, n := range []int{0, 1, 3, 10, 100, 1000} {
		b.Run(fmt.Sprintf("apps=%d", n), func(b *testing.B) {
			w, counted := benchWorld(n)
			ctx := context.Background()
			step := 0
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				step++
				w.svc.now = func() time.Time { return t0.Add(time.Duration(step) * jobInterval) }
				w.settings.jobs[0].Status = base.SettingStatusActive
				data := &queue.PeriodicExecData{Task: &entity.Task{ID: "task"}}
				if err := w.svc.Run(ctx, data); err != nil {
					b.Fatal(err)
				}
			}
			runs := float64(b.N)
			b.ReportMetric(float64(counted.db)/runs, "db/run")
			b.ReportMetric(float64(counted.docker)/runs, "docker/run")
			b.ReportMetric(float64(counted.loads)/runs, "logq/run")
			b.ReportMetric(float64(counted.redis)/runs, "redis/run")
			b.ReportMetric(float64(counted.scaled)/runs, "scaled/run")
		})
	}
}
