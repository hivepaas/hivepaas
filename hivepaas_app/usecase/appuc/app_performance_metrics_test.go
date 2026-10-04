package appuc

import (
	"context"
	"net/netip"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/obi"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
	"github.com/hivepaas/hivepaas/services/docker"
	"github.com/hivepaas/hivepaas/services/logging"
)

// histogram is n requests, all within the bound of le, f of them failed.
func histogram(n, failed int64, le string) logging.OBIHistogram {
	buckets := map[string]int64{"leInf": n}
	for _, field := range obi.BucketFields() {
		if field == le {
			le = "" // from here on, every one
		}
		if le == "" {
			buckets[field] = n
		}
	}
	return logging.OBIHistogram{Count: n, Errors: failed, SumMs: float64(n), Buckets: buckets}
}

// A step without a row is a point with none; a step's quantiles are its
// buckets'.
func TestEveryPerformanceStep(t *testing.T) {
	at := time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)
	w := window{start: at, end: at.Add(3 * time.Minute), step: time.Minute}
	got := everyPerformanceStep([]*logging.OBIBucket{
		{Time: at.Add(time.Minute), OBIHistogram: histogram(4, 1, "le10")},
	}, w)
	if assert.Len(t, got, 3) {
		assert.Equal(t, int64(0), got[0].Requests)
		assert.Nil(t, got[0].P50)
		assert.Equal(t, int64(4), got[1].Requests)
		assert.Equal(t, int64(1), got[1].Errors)
		assert.InDelta(t, 7.5, *got[1].P50, 1e-9, "within (5, 10]")
		assert.Equal(t, at.Add(2*time.Minute), got[2].Time)
	}
}

// The calls by kind, the busiest first, each a point per step.
func TestDependencyKinds(t *testing.T) {
	at := time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)
	w := window{start: at, end: at.Add(2 * time.Minute), step: time.Minute}
	got := dependencyKinds([]*logging.OBIBucket{
		{Time: at, Keys: map[string]string{"kind": "http"}, OBIHistogram: histogram(2, 0, "le50")},
		{Time: at, Keys: map[string]string{"kind": "db"}, OBIHistogram: histogram(5, 0, "le5")},
		{Time: at.Add(time.Minute), Keys: map[string]string{"kind": "http"}, OBIHistogram: histogram(1, 1, "le50")},
	}, w)
	if assert.Len(t, got, 2) {
		assert.Equal(t, "db", got[0].Kind)
		assert.Equal(t, int64(5), got[0].Totals.Requests)
		assert.Equal(t, "http", got[1].Kind)
		assert.Equal(t, int64(3), got[1].Totals.Requests)
		assert.Equal(t, int64(1), got[1].Totals.Errors)
		assert.Len(t, got[1].Series, 2)
		assert.Equal(t, int64(0), got[0].Series[1].Requests)
	}
}

// Operations are merged into their peers, whose numbers are their sum - the
// quantiles from the summed buckets - the busiest peer first.
func TestDependencyPeers(t *testing.T) {
	db := &entity.App{ID: "pg", Key: "pg", Name: "Postgres"}
	resolve := func(kind, peer string) *entity.App {
		if peer == "postgresql/shop" {
			return db
		}
		return nil
	}
	got := dependencyPeers([]*logging.OBIGroup{
		{Keys: map[string]string{"kind": "db", "peer": "postgresql/shop", "operation": "SELECT"},
			OBIHistogram: histogram(6, 0, "le5")},
		{Keys: map[string]string{"kind": "http", "peer": "api.example.com:443", "method": "GET"},
			OBIHistogram: histogram(9, 3, "le250")},
		{Keys: map[string]string{"kind": "db", "peer": "postgresql/shop", "operation": "INSERT"},
			OBIHistogram: histogram(4, 1, "le25")},
	}, resolve)
	if !assert.Len(t, got, 2) {
		return
	}
	pg := got[0]
	assert.Equal(t, "postgresql/shop", pg.Peer)
	assert.Equal(t, &appdto.AppDependencyAppResp{ID: "pg", Key: "pg", Name: "Postgres"}, pg.App)
	assert.Equal(t, int64(10), pg.Requests)
	assert.Equal(t, int64(1), pg.Errors)
	assert.InDelta(t, 25.0/6, *pg.P50, 1e-9, "the 5th of 10, 6 of them within 5 ms")
	assert.InDelta(t, 23.125, *pg.P95, 1e-9, "the 9.5th of 10 between 10 and 25 ms")
	if assert.Len(t, pg.Operations, 2) {
		assert.Equal(t, "SELECT", pg.Operations[0].Operation)
		assert.Equal(t, int64(4), pg.Operations[1].Requests)
	}
	assert.Equal(t, "api.example.com:443", got[1].Peer)
	assert.Nil(t, got[1].App)
	assert.Equal(t, "GET", got[1].Operations[0].Method)
}

// A peer is the env's app by its key, its service's name, or an address of
// its today; a database by its engine and database, a cache by its engine
// when it is the only.
func TestPeerResolver(t *testing.T) {
	api := &entity.App{ID: "api", Key: "api", GlobalKey: "p1_dev_api"}
	pg := &entity.App{ID: "pg"}
	pg2 := &entity.App{ID: "pg2"}
	redis := &entity.App{ID: "redis"}
	r := &peerResolver{
		byName: map[string]*entity.App{"api": api, "p1_dev_api": api, "tasks.p1_dev_api": api},
		byIP:   map[netip.Addr]*entity.App{netip.MustParseAddr("10.0.1.7"): api},
		dbs: []*peerDatabase{{app: pg, engine: "postgres", dbName: "shop"},
			{app: pg2, engine: "postgres", dbName: "crm"}, {app: redis, engine: "redis"}},
	}
	for peer, want := range map[string]*entity.App{
		"api:8080": api, "API": api, "p1_dev_api:80": api, "tasks.p1_dev_api:80": api, "api.": api,
		"10.0.1.7:8080": api, "10.0.1.8:8080": nil, "api.example.com:443": nil, "": nil,
	} {
		assert.Equal(t, want, r.resolve(obi.KindHTTP, peer), peer)
	}
	assert.Equal(t, pg, r.resolve(obi.KindDB, "postgresql/shop"))
	assert.Equal(t, pg2, r.resolve(obi.KindDB, "postgresql/crm"))
	assert.Nil(t, r.resolve(obi.KindDB, "postgresql/other"), "another database: an outside one")
	assert.Nil(t, r.resolve(obi.KindDB, "postgresql"), "no database named: either")
	assert.Equal(t, redis, r.resolve(obi.KindDB, "redis/0"))
	assert.Nil(t, r.resolve(obi.KindDB, "mysql/shop"))
}

// coverageDocker answers the app's tasks.
type coverageDocker struct {
	docker.Manager
	tasks []swarm.Task
}

func (d *coverageDocker) ServiceTaskList(context.Context, string, []swarm.TaskState,
	...docker.TaskListOption) (*client.TaskListResult, error) {
	return &client.TaskListResult{Items: d.tasks}, nil
}

// coverageLogging answers the nodes' statuses.
type coverageLogging struct {
	loggingservice.Service
	statuses map[string]*loggingservice.PerformanceNodeStatus
}

func (l *coverageLogging) PerformanceStatus(context.Context, database.IDB) (
	map[string]*loggingservice.PerformanceNodeStatus, error) {
	return l.statuses, nil
}

// The nodes an app runs on now, those of them that run OBI by the settings
// and can by their agents; the reason when none does.
func TestPerformanceCoverage(t *testing.T) {
	task := func(node string, state swarm.TaskState) swarm.Task {
		return swarm.Task{NodeID: node, Status: swarm.TaskStatus{State: state}}
	}
	status := func(ok bool, reasons ...string) *loggingservice.PerformanceNodeStatus {
		return &loggingservice.PerformanceNodeStatus{Status: obi.Status{Preflight: obi.Preflight{OK: ok,
			Reasons: reasons}}}
	}
	perf := &entity.LoggingPerformance{Enabled: true, Nodes: []*entity.LoggingPerformanceNode{{ID: "n1"}, {ID: "n2"}}}
	cover := func(tasks []swarm.Task, statuses map[string]*loggingservice.PerformanceNodeStatus) (
		appdto.AppPerformanceMetricsHeadResp, error) {
		uc := &UC{dockerManager: &coverageDocker{tasks: tasks}, loggingService: &coverageLogging{statuses: statuses}}
		var head appdto.AppPerformanceMetricsHeadResp
		err := uc.performanceCoverage(context.Background(), &entity.App{ServiceID: "s"}, perf, &head)
		return head, err
	}

	head, err := cover([]swarm.Task{task("n1", swarm.TaskStateRunning), task("n3", swarm.TaskStateRunning),
		task("n2", swarm.TaskStateShutdown)}, map[string]*loggingservice.PerformanceNodeStatus{"n1": status(true)})
	assert.NoError(t, err)
	assert.Equal(t, appdto.AppPerformanceMetricsHeadResp{Nodes: 2, NodesCovered: 1}, head)

	head, _ = cover([]swarm.Task{task("n3", swarm.TaskStateRunning)}, nil)
	assert.Equal(t, performanceReasonNodeDisabled, head.Reason)

	head, _ = cover([]swarm.Task{task("n1", swarm.TaskStateRunning), task("n2", swarm.TaskStateRunning)},
		map[string]*loggingservice.PerformanceNodeStatus{"n1": status(false, "no-btf"),
			"n2": status(false, "lockdown", "no-btf")})
	assert.Equal(t, performanceReasonNodeUnsupported, head.Reason)
	assert.Equal(t, []string{"lockdown", "no-btf"}, head.PreflightReasons)

	head, _ = cover([]swarm.Task{task("n2", swarm.TaskStateRunning)}, nil)
	assert.Equal(t, 1, head.NodesCovered, "a node whose agent said nothing yet may run it")

	head, _ = cover(nil, nil)
	assert.Equal(t, appdto.AppPerformanceMetricsHeadResp{}, head, "running nowhere: its past is shown")
}

// perfApps loads one app, with the feature settings given.
type perfApps struct {
	appservice.Service
	features *entity.AppFeatureSettings
}

func (s *perfApps) LoadAppWithFeatureSettings(context.Context, database.IDB, string, string, bool, bool,
	...bunex.SelectQueryOption) (*entity.App, *entity.AppFeatureSettings, error) {
	return &entity.App{ID: "a1"}, s.features, nil
}

// perfSettings answers the logging settings given; none when nil.
type perfSettings struct {
	repository.SettingRepo
	logging *entity.LoggingSettings
}

func (s *perfSettings) GetSingle(context.Context, database.IDB, *entity.ObjectScope, base.SettingType, bool,
	...bunex.SelectQueryOption) (*entity.Setting, error) {
	if s.logging == nil {
		return nil, hperrors.NewNotFound("Setting")
	}
	setting := &entity.Setting{Type: base.SettingTypeLogging}
	setting.MustSetData(s.logging)
	return setting, nil
}

// While routes and calls are off, for the system or the app, the reason is
// answered from the settings alone: Docker, the agent and the logs are not
// asked - they are nil here, and asking them would fail.
func TestPerformanceViewWhileOffAsksNothingElse(t *testing.T) {
	appOn := &entity.AppFeatureSettings{PerformanceSettings: &entity.AppFeaturePerformanceSettings{Enabled: true}}
	systemOn := &entity.LoggingSettings{Enabled: true, Performance: &entity.LoggingPerformance{Enabled: true}}
	for name, c := range map[string]struct {
		logging  *entity.LoggingSettings
		features *entity.AppFeatureSettings
		reason   string
	}{
		"never saved": {nil, appOn, performanceReasonDisabled},
		"system off":  {&entity.LoggingSettings{Enabled: true}, appOn, performanceReasonDisabled},
		"app off":     {systemOn, &entity.AppFeatureSettings{}, performanceReasonAppDisabled},
	} {
		uc := &UC{appService: &perfApps{features: c.features}, settingRepo: &perfSettings{logging: c.logging}}
		view, err := uc.loadPerformanceView(context.Background(),
			&appdto.GetAppPerformanceMetricsReq{ProjectID: "p1", AppID: "a1", Range: "1h"})
		if assert.NoError(t, err, name) {
			assert.Equal(t, c.reason, view.head.Reason, name)
			assert.False(t, view.head.Available, name)
		}
	}
}
