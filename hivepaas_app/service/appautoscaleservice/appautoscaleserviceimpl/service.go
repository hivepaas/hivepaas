package appautoscaleserviceimpl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appautoscaleservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemeventbusservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
	logs "github.com/hivepaas/hivepaas/services/logging"
)

const (
	// jobInterval is how often the apps are looked at: the periodic base
	// interval. It does not slow down when they rest - the next burst is then.
	jobInterval = 15 * time.Second
	jobTimeout  = time.Minute
	jobName     = "App autoscale"

	// window is how far back a run reads the load: long enough to hold most
	// of the calls of a function with the default 30 s timeout, and 4 of the
	// agent's rows.
	window = time.Minute

	// logLag is how far behind now the window ends: the lines of the last
	// seconds may not have reached VictoriaLogs yet - later from a node further
	// away, a Traefik replica's or a function's - and missing, they would read
	// as less load.
	logLag = 10 * time.Second

	// shortWindow is the window's last part, whose requests and calls are read
	// apart in the same query: a burst the minute's average takes most of a
	// minute to see is acted on in one run.
	shortWindow = 15 * time.Second

	// stateTTL keeps an app's state between runs, and lets it go once the job
	// has stopped running for it.
	stateTTL      = time.Hour
	stateKeyFmt   = "autoscale:app:%s"
	updateRetries = 2

	// clusterFullAfter is how long a service may run fewer tasks than it
	// wants before autoscale takes it that the cluster has no room for more.
	clusterFullAfter = 2 * time.Minute

	// legacyJobKind is the job's kind before it scaled every app: the setting
	// made then is renamed.
	legacyJobKind = "function-autoscale"
)

// stateStore keeps an app's state between runs.
type stateStore interface {
	load(ctx context.Context, appID string) state
	save(ctx context.Context, appID string, st state)
}

type service struct {
	db             database.IDB
	states         stateStore
	dockerManager  docker.Manager
	traefikService traefikservice.Service
	hpAppService   hpappservice.Service
	appRepo        repository.AppRepo
	settingRepo    repository.SettingRepo
	taskRepo       repository.TaskRepo
	loggingService loggingservice.Service
	systemEventBus systemeventbusservice.Service
	logger         logging.Logger
	now            func() time.Time
	// inTx runs in a transaction of the store's.
	inTx func(ctx context.Context, exec func(db database.IDB) error) error
}

func New(
	db *database.DB,
	redisClient rediscache.Client,
	dockerManager docker.Manager,
	traefikService traefikservice.Service,
	hpAppService hpappservice.Service,
	appRepo repository.AppRepo,
	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,
	loggingService loggingservice.Service,
	systemEventBus systemeventbusservice.Service,
	logger logging.Logger,
) appautoscaleservice.Service {
	return &service{db: db, states: redisStates{client: redisClient}, dockerManager: dockerManager,
		traefikService: traefikService, hpAppService: hpAppService, appRepo: appRepo,
		settingRepo: settingRepo, taskRepo: taskRepo, loggingService: loggingService, systemEventBus: systemEventBus,
		logger: logger, now: timeutil.NowUTC,
		inTx: func(ctx context.Context, exec func(db database.IDB) error) error {
			return hperrors.Wrap(transaction.Execute(ctx, db, func(tx database.Tx) error { return exec(tx) }))
		}}
}

// enabledAutoscales are the autoscale settings that are on, by app id.
func (s *service) enabledAutoscales(ctx context.Context, db database.IDB) (map[string]*entity.AppAutoscale, error) {
	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppAutoscale),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.data->>'enabled' = 'true'"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	out := make(map[string]*entity.AppAutoscale, len(settings))
	for _, setting := range settings {
		autoscale, err := setting.AsAppAutoscale()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		out[setting.ObjectID] = autoscale
	}
	return out, nil
}

func (s *service) EnsureJob(ctx context.Context, db database.IDB) error {
	autoscales, err := s.enabledAutoscales(ctx, db)
	if err != nil {
		return err
	}
	status := gofn.If(len(autoscales) > 0, base.SettingStatusActive, base.SettingStatusDisabled)

	// Locked, as stopIdleJob locks it: one of the two waits for the other,
	// and reads what it wrote.
	jobs, err := s.jobSettings(ctx, db, bunex.SelectFor("UPDATE OF setting"))
	if err != nil {
		return err
	}
	now := s.now()
	switch {
	case len(jobs) == 0 && status == base.SettingStatusDisabled:
		return nil
	case len(jobs) == 0:
		job := &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.ObjectScopeGlobal,
			Type:      base.SettingTypePeriodicJob,
			Kind:      string(base.PeriodicKindAppAutoscale),
			Status:    status,
			Name:      jobName,
			Version:   entity.CurrentPeriodicJobVersion,
			CreatedAt: now,
			UpdatedAt: now,
		}
		job.MustSetData(&entity.PeriodicJob{
			Interval: timeutil.Duration(jobInterval), Timeout: timeutil.Duration(jobTimeout)})
		if err = s.settingRepo.Insert(ctx, db, job); err != nil {
			return hperrors.Wrap(err)
		}
	case jobs[0].Status == status && jobs[0].Kind == string(base.PeriodicKindAppAutoscale):
		return nil
	default:
		// Its status, and its kind and name when made before it scaled every
		// app: the executor knows it by the kind.
		job := jobs[0]
		job.Status = status
		job.Kind = string(base.PeriodicKindAppAutoscale)
		job.Name = jobName
		job.UpdateVer++
		job.UpdatedAt = now
		err = s.settingRepo.Update(ctx, db, job,
			bunex.UpdateColumns("status", "kind", "name", "update_ver", "updated_at"))
		if err != nil {
			return hperrors.Wrap(err)
		}
	}
	// The workers keep the periodic jobs in a cache: tell them it moved.
	_ = s.systemEventBus.Publish(ctx, base.SystemEventPeriodicSettingsReload)
	return nil
}

// jobSettings is the job's setting, made once, whatever its status: none
// before an app first had autoscale on. Found by its kind before it scaled
// every app too.
func (s *service) jobSettings(
	ctx context.Context, db database.IDB, opts ...bunex.SelectQueryOption,
) ([]*entity.Setting, error) {
	jobs, _, err := s.settingRepo.List(ctx, db, nil, nil, append([]bunex.SelectQueryOption{
		bunex.SelectWhere("setting.type = ?", base.SettingTypePeriodicJob),
		bunex.SelectWhereIn("setting.kind IN (?)", string(base.PeriodicKindAppAutoscale), legacyJobKind),
		bunex.SelectWhere("setting.scope = ?", base.ObjectScopeGlobal),
		bunex.SelectLimit(1),
	}, opts...)...)
	return jobs, hperrors.Wrap(err)
}

// stopIdleJob turns the job off when no app has autoscale on any more - the
// last one's app deleted, or its project, which turns no setting off. The
// job's row is locked before the apps are looked at, as EnsureJob locks it:
// an app turning autoscale on meanwhile is seen here, or turns the job on
// again after.
func (s *service) stopIdleJob(ctx context.Context) error {
	stopped := false
	err := s.inTx(ctx, func(db database.IDB) error {
		jobs, err := s.jobSettings(ctx, db, bunex.SelectFor("UPDATE OF setting"))
		if err != nil || len(jobs) == 0 || jobs[0].Status != base.SettingStatusActive {
			return err
		}
		autoscales, err := s.enabledAutoscales(ctx, db)
		if err != nil || len(autoscales) > 0 {
			return err
		}
		job := jobs[0]
		job.Status = base.SettingStatusDisabled
		job.UpdateVer++
		job.UpdatedAt = s.now()
		stopped = true
		return hperrors.Wrap(s.settingRepo.Update(ctx, db, job,
			bunex.UpdateColumns("status", "update_ver", "updated_at")))
	})
	if err != nil {
		return err
	}
	if stopped {
		// The workers keep the periodic jobs in a cache: tell them it stopped.
		_ = s.systemEventBus.Publish(ctx, base.SystemEventPeriodicSettingsReload)
	}
	return nil
}

func (s *service) Events(
	ctx context.Context, db database.IDB, appID string, since time.Time, limit int,
) ([]*appautoscaleservice.Event, error) {
	jobs, err := s.jobSettings(ctx, db)
	if err != nil || len(jobs) == 0 {
		return nil, err
	}
	// The job's tasks are a run each, listing the functions it scaled: the
	// function's are those whose output names it. The pattern holds the app
	// alone - the output's other fields are not to be matched.
	match, err := json.Marshal(map[string]any{"scaled": []map[string]string{{"app": appID}}})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	opts := []bunex.SelectQueryOption{
		bunex.SelectWhere("task.run_at >= ?", since),
		bunex.SelectWhere("task.output::jsonb @> ?::jsonb", string(match)),
		bunex.SelectOrder("task.run_at DESC"),
	}
	if limit > 0 {
		opts = append(opts, bunex.SelectLimit(limit))
	}
	tasks, _, err := s.taskRepo.ListByTarget(ctx, db, jobs[0].ID, nil, opts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	events := make([]*appautoscaleservice.Event, 0, len(tasks))
	for _, task := range tasks {
		out := &entity.TaskAppAutoscaleOutput{}
		if err := json.Unmarshal([]byte(task.Output), out); err != nil {
			continue
		}
		for _, change := range out.Scaled {
			if change.App != appID {
				continue
			}
			events = append(events, &appautoscaleservice.Event{
				Time: task.RunAt, From: change.From, To: change.To, InFlight: change.InFlight,
				Calls: change.Calls, Throttled: change.Throttled, Requests: change.Requests, CPU: change.CPU,
				Reason: change.Reason,
			})
		}
	}
	return events, nil
}

func (s *service) Run(ctx context.Context, data *queue.PeriodicExecData) error {
	autoscales, err := s.enabledAutoscales(ctx, s.db)
	if err != nil {
		return err
	}
	if len(autoscales) == 0 {
		return s.stopIdleJob(ctx)
	}
	ids := gofn.MapKeys(autoscales)
	sort.Strings(ids)
	apps, err := s.appRepo.ListByIDs(ctx, s.db, "", ids)
	if err != nil {
		return hperrors.Wrap(err)
	}
	concurrency, err := s.concurrencies(ctx, ids)
	if err != nil {
		return err
	}
	services, err := s.services(ctx, apps)
	if err != nil {
		return err
	}

	now := s.now()
	r := &runData{now: now, autoscales: autoscales, concurrency: concurrency, services: services,
		end: now.Add(-logLag)}
	r.start, r.shortStart = r.end.Add(-window), r.end.Add(-shortWindow)
	var errs []error
	// No data is not no load: what cannot be read moves nothing.
	if err = s.readFunctions(ctx, r, apps); err != nil {
		errs = append(errs, err)
	}
	if err = s.readApps(ctx, r, apps); err != nil {
		errs = append(errs, err)
	}

	output := &entity.TaskAppAutoscaleOutput{}
	for _, app := range apps {
		scaled, err := s.runApp(ctx, r, app)
		if err != nil {
			s.logger.Warnf("autoscale: %s: %v", app.Name, err)
			continue
		}
		if scaled != nil {
			output.Scaled = append(output.Scaled, scaled)
		}
	}
	err = errors.Join(errs...)
	if len(output.Scaled) > 0 {
		// A run that changed nothing leaves no task: one every 15 s would
		// bury the ones that did. The queue saves it as it is: its status and
		// end are the run's to set.
		data.Task.MustSetOutput(output)
		data.Task.Status = gofn.If(err == nil, base.TaskStatusDone, base.TaskStatusFailed)
		if data.Task.StartedAt.IsZero() {
			data.Task.StartedAt = now
		}
		data.Task.EndedAt = s.now()
		data.SaveTask = true
	}
	return hperrors.Wrap(err)
}

// runData is what one run reads, for every app at once, over [start, end) -
// and its last part, from shortStart, apart.
type runData struct {
	now                    time.Time
	start, end, shortStart time.Time
	autoscales             map[string]*entity.AppAutoscale
	concurrency            map[string]int
	services               map[string]*swarm.Service

	functionsRead bool
	functionLoads map[string]*logs.InvocationLoad

	exposed      map[string]bool
	requestsRead bool
	requests     map[string]*logs.RequestLoad
	cpuRead      bool
	cpu          map[string][]*logs.ContainerCPU
}

// services are the apps' services, with their tasks running and wanted, by
// id: one list for every app.
func (s *service) services(ctx context.Context, apps []*entity.App) (map[string]*swarm.Service, error) {
	out := make(map[string]*swarm.Service, len(apps))
	ids := make([]string, 0, len(apps))
	for _, app := range apps {
		if app.ServiceID != "" {
			ids = append(ids, app.ServiceID)
		}
	}
	if len(ids) == 0 {
		return out, nil
	}
	list, err := s.dockerManager.ServiceList(ctx, func(opts *client.ServiceListOptions) {
		opts.Status = true
		for _, id := range ids {
			docker.FilterAdd(&opts.Filters, "id", id)
		}
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for i := range list.Items {
		out[list.Items[i].ID] = &list.Items[i]
	}
	return out, nil
}

// readFunctions reads the functions' calls, one query for them all.
func (s *service) readFunctions(ctx context.Context, r *runData, apps []*entity.App) error {
	ids := make([]string, 0, len(apps))
	for _, app := range apps {
		if r.concurrency[app.ID] > 0 {
			ids = append(ids, app.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	loads, err := s.loggingService.FunctionLoad(ctx, s.db, ids, r.start, r.end, r.shortStart)
	if err != nil {
		return hperrors.Wrap(err)
	}
	r.functionsRead, r.functionLoads = true, loads
	return nil
}

// readApps reads the other apps' requests and CPU, one query for each signal,
// when the proxy and the agent mark their lines.
func (s *service) readApps(ctx context.Context, r *runData, apps []*entity.App) error {
	var onRequests, onCPU []string
	for _, app := range apps {
		a := r.autoscales[app.ID]
		if r.concurrency[app.ID] > 0 || a == nil {
			continue
		}
		if a.RequestsTarget > 0 {
			onRequests = append(onRequests, app.ID)
		}
		if a.CPUTarget > 0 {
			onCPU = append(onCPU, app.ID)
		}
	}
	var errs []error
	if len(onRequests) > 0 {
		if err := s.readRequests(ctx, r, onRequests); err != nil {
			errs = append(errs, err)
		}
	}
	if len(onCPU) > 0 {
		if err := s.readCPU(ctx, r, onCPU); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (s *service) readRequests(ctx context.Context, r *runData, ids []string) error {
	if reason, err := s.accessLogReadiness(ctx); err != nil || reason != "" {
		return err
	}
	exposed, err := s.exposed(ctx, ids)
	if err != nil {
		return err
	}
	loads, err := s.loggingService.RequestLoad(ctx, s.db, ids, r.start, r.end, r.shortStart)
	if err != nil {
		return hperrors.Wrap(err)
	}
	r.requestsRead, r.requests, r.exposed = true, loads, exposed
	return nil
}

func (s *service) readCPU(ctx context.Context, r *runData, ids []string) error {
	if reason, err := s.agentReadiness(ctx); err != nil || reason != "" {
		return err
	}
	loads, err := s.loggingService.CPULoad(ctx, s.db, ids, r.start, r.end)
	if err != nil {
		return hperrors.Wrap(err)
	}
	r.cpuRead, r.cpu = true, loads
	return nil
}

// concurrencies are the functions' Concurrency, by app id, from their
// deployment settings: an app not in it is not a function.
func (s *service) concurrencies(ctx context.Context, ids []string) (map[string]int, error) {
	settings, _, err := s.settingRepo.List(ctx, s.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
		bunex.SelectWhereIn("setting.object_id IN (?)", ids...),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	out := make(map[string]int, len(settings))
	for _, setting := range settings {
		deployment, err := setting.AsAppDeploymentSettings()
		if err != nil || deployment.ActiveMethod != base.DeploymentMethodFunction || deployment.FunctionSource == nil {
			continue
		}
		out[setting.ObjectID] = gofn.Coalesce(deployment.FunctionSource.MaxConcurrency,
			base.FunctionMaxConcurrencyDefault)
	}
	return out, nil
}

// exposed is whether each app is reached by a domain through the proxy, by
// app id.
func (s *service) exposed(ctx context.Context, ids []string) (map[string]bool, error) {
	settings, _, err := s.settingRepo.List(ctx, s.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppRouting),
		bunex.SelectWhereIn("setting.object_id IN (?)", ids...),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	out := make(map[string]bool, len(settings))
	for _, setting := range settings {
		out[setting.ObjectID] = isExposed(setting)
	}
	return out, nil
}

// runApp decides one app's replicas and applies them; the change, or nil for
// none.
func (s *service) runApp(ctx context.Context, r *runData, app *entity.App) (*entity.AppAutoscaleChange, error) {
	autoscale := r.autoscales[app.ID]
	if autoscale == nil || app.ServiceID == "" || app.Status != base.AppStatusActive {
		return nil, nil
	}
	svc := r.services[app.ServiceID]
	if svc == nil {
		return nil, nil
	}
	concurrency := r.concurrency[app.ID]
	isFunction := concurrency > 0
	current, ok := scalable(svc, app.ID, isFunction)
	if !ok {
		return nil, nil
	}

	change := &entity.AppAutoscaleChange{App: app.ID, Name: app.Name, From: current}
	var a ask
	if isFunction {
		if !r.functionsRead {
			return nil, nil
		}
		load := r.functionLoads[app.ID]
		in := &input{Current: current, Concurrency: concurrency, Target: autoscale.Target, Window: window,
			ShortWindow: shortWindow, Load: load}
		a, change.InFlight = askFunction(in)
		change.Calls, change.Throttled = callsOf(load), throttledOf(load)
		change.Concurrency, change.Target = concurrency, autoscale.Target
	} else {
		in := s.appInputOf(r, app, autoscale, svc, current)
		var read bool
		if a, change.InFlight, read = askApp(in); !read {
			return nil, nil
		}
		if in.RequestsRead {
			change.RequestsTarget = autoscale.RequestsTarget
			if in.Requests != nil {
				change.Requests = in.Requests.Requests
			}
		}
		if utilization, ok := cpuUtilization(in.CPU, in.Reservation); ok && in.CPURead {
			change.CPU, change.CPUTarget = math.Round(utilization*1000)/10, autoscale.CPUTarget //nolint:mnd // %
		}
	}

	st := s.states.load(ctx, app.ID)
	st.ShortSince = shortSince(svc, st.ShortSince, r.now)
	if st.ShortSince > 0 && r.now.Sub(time.Unix(st.ShortSince, 0)) >= clusterFullAfter && a.Want > current {
		// The cluster has no room for what it scaled to: no more until it has.
		a.Want, a.Urgent = current, false
	}
	d, next := settle(bounds{Current: current, Min: autoscale.MinReplicas, Max: autoscale.MaxReplicas,
		ScaleInDelay: autoscale.ScaleInDelay.ToDuration(), Now: r.now}, a, st)
	s.states.save(ctx, app.ID, next)
	if d.Desired == current {
		return nil, nil
	}

	desired := uint64(d.Desired) //nolint:gosec // between Min and Max, 1 to 50
	err := s.dockerManager.ServiceUpdateFunc(ctx, svc.ID, svc, func(_ int, svc *swarm.Service) (bool, error) {
		// Read again: stopped, or scaled by hand, meanwhile - leave it.
		if now, ok := scalable(svc, app.ID, isFunction); !ok || now != current {
			return false, nil
		}
		svc.Spec.Mode.Replicated.Replicas = &desired
		return true, nil
	}, updateRetries, 0)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	change.To, change.Reason = d.Desired, d.Reason
	return change, nil
}

// appInputOf is what an app other than a function is decided from this run.
func (s *service) appInputOf(
	r *runData, app *entity.App, autoscale *entity.AppAutoscale, svc *swarm.Service, current int,
) *appInput {
	in := &appInput{Current: current, Window: window, ShortWindow: shortWindow,
		RequestsTarget: autoscale.RequestsTarget, CPUTarget: autoscale.CPUTarget,
		Reservation: cpuReservation(&svc.Spec)}
	if r.requestsRead && r.exposed[app.ID] {
		in.RequestsRead, in.Requests = true, r.requests[app.ID]
	}
	// No row of a running app's is the agent not reporting it, not an idle
	// app: it holds.
	if r.cpuRead && hasAppLabel(&svc.Spec, app.ID) && len(r.cpu[app.ID]) > 0 {
		in.CPURead, in.CPU = true, r.cpu[app.ID]
	}
	return in
}

// shortSince is since when a service has run fewer tasks than it wants, 0
// when it runs them all.
func shortSince(svc *swarm.Service, since int64, now time.Time) int64 {
	status := svc.ServiceStatus
	if status == nil || status.RunningTasks >= status.DesiredTasks {
		return 0
	}
	if since == 0 {
		return now.Unix()
	}
	return since
}

// scalable is a service's replicas when autoscale may change them: replicated,
// running - 0 is how HivePaaS stops an app - not mid-update, and no port
// published on its node. A function's calls must be in its logs, or no load
// would read as none.
func scalable(svc *swarm.Service, appID string, isFunction bool) (int, bool) {
	mode := svc.Spec.Mode.Replicated
	if mode == nil || mode.Replicas == nil || *mode.Replicas == 0 {
		return 0, false
	}
	if u := svc.UpdateStatus; u != nil &&
		(u.State == swarm.UpdateStateUpdating || u.State == swarm.UpdateStateRollbackStarted) {
		return 0, false
	}
	if hasHostPorts(&svc.Spec) {
		return 0, false
	}
	if isFunction &&
		!appservice.HasLogIdentity(svc.Spec.TaskTemplate.LogDriver, svc.Spec.TaskTemplate.ContainerSpec, appID) {
		return 0, false
	}
	return int(*mode.Replicas), true //nolint:gosec // a service's replicas
}

// redisStates keeps the states in Redis, shared by the workers.
type redisStates struct {
	client rediscache.Client
}

func (r redisStates) load(ctx context.Context, appID string) state {
	var st state
	raw, err := r.client.Get(ctx, fmt.Sprintf(stateKeyFmt, appID)).Bytes()
	if err == nil {
		_ = json.Unmarshal(raw, &st)
	}
	return st
}

// save keeps it for the next run; lost, it starts again, which only delays a
// scale-out by a run or a scale-in by its delay.
func (r redisStates) save(ctx context.Context, appID string, st state) {
	key := fmt.Sprintf(stateKeyFmt, appID)
	if st == (state{}) {
		_ = r.client.Del(ctx, key).Err()
		return
	}
	raw, _ := json.Marshal(st)
	_ = r.client.Set(ctx, key, raw, stateTTL).Err()
}

func callsOf(load *logs.InvocationLoad) int64 {
	if load == nil {
		return 0
	}
	return load.Calls
}

func throttledOf(load *logs.InvocationLoad) int64 {
	if load == nil {
		return 0
	}
	return load.Throttled
}
