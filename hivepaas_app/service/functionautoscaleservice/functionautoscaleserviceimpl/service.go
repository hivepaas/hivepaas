package functionautoscaleserviceimpl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionautoscaleservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemeventbusservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/services/docker"
	logs "github.com/hivepaas/hivepaas/services/logging"
)

const (
	// jobInterval is how often the functions are looked at: the periodic base
	// interval. It does not slow down when they rest - the next burst is then.
	jobInterval = 15 * time.Second
	jobTimeout  = time.Minute
	jobName     = "Function autoscale"

	// window is how far back a run reads the calls: long enough to hold most
	// of the calls of a function with the default 30 s timeout.
	window = time.Minute

	// stateTTL keeps a function's state between runs, and lets it go once the
	// job has stopped running for it.
	stateTTL      = time.Hour
	stateKeyFmt   = "autoscale:function:%s"
	updateRetries = 2
)

// stateStore keeps a function's state between runs.
type stateStore interface {
	load(ctx context.Context, appID string) state
	save(ctx context.Context, appID string, st state)
}

type service struct {
	db             database.IDB
	states         stateStore
	dockerManager  docker.Manager
	appRepo        repository.AppRepo
	settingRepo    repository.SettingRepo
	loggingService loggingservice.Service
	systemEventBus systemeventbusservice.Service
	logger         logging.Logger
	now            func() time.Time
}

func New(
	db *database.DB,
	redisClient rediscache.Client,
	dockerManager docker.Manager,
	appRepo repository.AppRepo,
	settingRepo repository.SettingRepo,
	loggingService loggingservice.Service,
	systemEventBus systemeventbusservice.Service,
	logger logging.Logger,
) functionautoscaleservice.Service {
	return &service{db: db, states: redisStates{client: redisClient}, dockerManager: dockerManager, appRepo: appRepo,
		settingRepo: settingRepo, loggingService: loggingService, systemEventBus: systemEventBus,
		logger: logger, now: timeutil.NowUTC}
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

	jobs, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypePeriodicJob),
		bunex.SelectWhere("setting.kind = ?", base.PeriodicKindFunctionAutoscale),
		bunex.SelectWhere("setting.scope = ?", base.ObjectScopeGlobal),
		bunex.SelectLimit(1),
	)
	if err != nil {
		return hperrors.Wrap(err)
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
			Kind:      string(base.PeriodicKindFunctionAutoscale),
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
	case jobs[0].Status == status:
		return nil
	default:
		job := jobs[0]
		job.Status = status
		job.UpdateVer++
		job.UpdatedAt = now
		if err = s.settingRepo.Update(ctx, db, job, bunex.UpdateColumns("status", "update_ver", "updated_at")); err != nil {
			return hperrors.Wrap(err)
		}
	}
	// The workers keep the periodic jobs in a cache: tell them it moved.
	_ = s.systemEventBus.Publish(ctx, base.SystemEventPeriodicSettingsReload)
	return nil
}

func (s *service) Run(ctx context.Context, data *queue.PeriodicExecData) error {
	autoscales, err := s.enabledAutoscales(ctx, s.db)
	if err != nil || len(autoscales) == 0 {
		return err
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

	now := s.now()
	// No data is not no load: when the calls cannot be read, nothing moves.
	loads, err := s.loggingService.FunctionLoad(ctx, s.db, ids, now.Add(-window), now)
	if err != nil {
		return hperrors.Wrap(err)
	}

	output := &entity.TaskFunctionAutoscaleOutput{}
	for _, app := range apps {
		scaled, err := s.runFunction(ctx, app, autoscales[app.ID], concurrency[app.ID], loads[app.ID], now)
		if err != nil {
			s.logger.Warnf("function autoscale: %s: %v", app.Name, err)
			continue
		}
		if scaled != nil {
			output.Scaled = append(output.Scaled, scaled)
		}
	}
	if len(output.Scaled) > 0 {
		// A run that changed nothing leaves no task: one every 15 s would
		// bury the ones that did.
		data.Task.MustSetOutput(output)
		data.SaveTask = true
	}
	return nil
}

// concurrencies are the functions' Concurrency, by app id, from their
// deployment settings.
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
		if err != nil || deployment.FunctionSource == nil {
			continue
		}
		out[setting.ObjectID] = gofn.Coalesce(deployment.FunctionSource.MaxConcurrency,
			base.FunctionMaxConcurrencyDefault)
	}
	return out, nil
}

// runFunction decides one function's replicas and applies them; the change,
// or nil for none.
func (s *service) runFunction(
	ctx context.Context,
	app *entity.App,
	autoscale *entity.AppAutoscale,
	concurrency int,
	load *logs.InvocationLoad,
	now time.Time,
) (*entity.FunctionAutoscaleChange, error) {
	if autoscale == nil || concurrency == 0 || app.ServiceID == "" || app.Status != base.AppStatusActive {
		return nil, nil
	}
	inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
	if errors.Is(err, hperrors.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	svc := &inspect.Service
	current, ok := scalable(svc, app.ID)
	if !ok {
		return nil, nil
	}

	st := s.states.load(ctx, app.ID)
	d, next := decide(&input{
		Current: current, Min: autoscale.MinReplicas, Max: autoscale.MaxReplicas, Concurrency: concurrency,
		Target: autoscale.Target, Window: window, ScaleInDelay: autoscale.ScaleInDelay.ToDuration(), Now: now,
		Load: load,
	}, st)
	s.states.save(ctx, app.ID, next)
	if d.Desired == current {
		return nil, nil
	}

	desired := uint64(d.Desired) //nolint:gosec // between Min and Max, 1 to 50
	err = s.dockerManager.ServiceUpdateFunc(ctx, svc.ID, svc, func(_ int, svc *swarm.Service) (bool, error) {
		// Read again: stopped, or scaled by hand, meanwhile - leave it.
		if now, ok := scalable(svc, app.ID); !ok || now != current {
			return false, nil
		}
		svc.Spec.Mode.Replicated.Replicas = &desired
		return true, nil
	}, updateRetries, 0)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &entity.FunctionAutoscaleChange{
		App: app.ID, Name: app.Name, From: current, To: d.Desired, InFlight: d.InFlight,
		Calls: callsOf(load), Throttled: throttledOf(load), Concurrency: concurrency, Target: autoscale.Target,
		Reason: d.Reason,
	}, nil
}

// scalable is a service's replicas when autoscale may change them: replicated,
// running - 0 is how HivePaaS stops an app - not mid-update, and with its
// calls in its logs, or no load would read as none.
func scalable(svc *swarm.Service, appID string) (int, bool) {
	mode := svc.Spec.Mode.Replicated
	if mode == nil || mode.Replicas == nil || *mode.Replicas == 0 {
		return 0, false
	}
	if u := svc.UpdateStatus; u != nil &&
		(u.State == swarm.UpdateStateUpdating || u.State == swarm.UpdateStateRollbackStarted) {
		return 0, false
	}
	if !appservice.HasLogIdentity(svc.Spec.TaskTemplate.LogDriver, svc.Spec.TaskTemplate.ContainerSpec, appID) {
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
