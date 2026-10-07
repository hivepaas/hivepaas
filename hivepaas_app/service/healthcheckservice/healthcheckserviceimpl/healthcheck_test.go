package healthcheckserviceimpl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/healthcheckservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobtriggerservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// stateStore keeps each check's state as redis does: what is read is a copy
// of what was set, not the object that was.
type stateStore struct {
	cacherepository.HealthcheckStateRepo
	states map[string]cacheentity.HealthcheckState
}

func (s *stateStore) Get(_ context.Context, id string) (*cacheentity.HealthcheckState, error) {
	state, ok := s.states[id]
	if !ok {
		return nil, hperrors.Wrap(hperrors.ErrNotFound)
	}
	return &state, nil
}

func (s *stateStore) Set(_ context.Context, id string, state *cacheentity.HealthcheckState, _ time.Duration) error {
	s.states[id] = *state
	return nil
}

// notifier has a notification for every result, and records those sent: the
// result each one told.
type notifier struct {
	notificationservice.Service
	told []bool
}

func (n *notifier) GetNotificationForEvent(context.Context, database.IDB, *entity.ObjectScope,
	*entity.BaseEventNotification, bool, *entity.RefObjects) (*entity.Notification, error) {
	return &entity.Notification{}, nil
}

func (n *notifier) BuildTitlePrefixForScope(*entity.ObjectScope) string { return "" }

func (n *notifier) NotifyForTaskResult(_ context.Context, _ database.IDB,
	data *notificationservice.TaskResultNotificationReq) (*notificationservice.TaskResultNotificationResp, error) {
	n.told = append(n.told, data.ActionSucceeded)
	return &notificationservice.TaskResultNotificationResp{
		SendTs: timeutil.NowUTC(), DeliveryMap: map[string]bool{"email": true},
	}, nil
}

type triggers struct {
	schedjobtriggerservice.Service
	fired []base.SchedJobTriggerEvent
}

func (f *triggers) Fire(_ context.Context, event base.SchedJobTriggerEvent, _ *entity.App,
	_ *schedjobtriggerservice.TriggerInfo) (*schedjobtriggerservice.FireResult, error) {
	f.fired = append(f.fired, event)
	return &schedjobtriggerservice.FireResult{}, nil
}

// checkFixture is a REST check of a server that answers as told: a status,
// or nothing at all until the request is given up.
type checkFixture struct {
	svc      *service
	states   *stateStore
	notifier *notifier
	triggers *triggers
	setting  *entity.Setting

	status atomic.Int32
	hang   atomic.Bool
}

func newCheckFixture(t *testing.T, interval, timeout time.Duration) *checkFixture {
	t.Helper()
	f := &checkFixture{
		states:   &stateStore{states: map[string]cacheentity.HealthcheckState{}},
		notifier: &notifier{},
		triggers: &triggers{},
	}
	f.status.Store(http.StatusOK)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f.hang.Load() {
			<-r.Context().Done()
			return
		}
		w.WriteHeader(int(f.status.Load()))
	}))
	t.Cleanup(server.Close)

	f.setting = &entity.Setting{ID: "check", Type: base.SettingTypePeriodicJob}
	assert.NoError(t, f.setting.SetData(&entity.PeriodicJob{
		Interval: timeutil.Duration(interval),
		Timeout:  timeutil.Duration(timeout),
		Healthcheck: &entity.PeriodicHealthcheck{
			HealthcheckType: base.HealthcheckTypeREST,
			REST:            &entity.HealthcheckREST{URL: server.URL, ReturnCode: []int{http.StatusOK}},
		},
		Notification: &entity.PeriodicNotification{
			BaseEventNotification: &entity.BaseEventNotification{SuccessUseDefault: true, FailureUseDefault: true},
		},
	}))
	f.svc = &service{healthcheckStateRepo: f.states, notificationService: f.notifier, schedJobTriggerService: f.triggers}
	// The notification's message links to the dashboard.
	config.SetCurrent(&config.Config{})
	t.Cleanup(func() { config.SetCurrent(nil) })
	return f
}

// run is one run of the check: how it ended, and whether it is to be saved.
func (f *checkFixture) run(t *testing.T) (base.TaskStatus, bool) {
	t.Helper()
	scope := entity.NewObjectScopeApp("app", "", "p", "dev")
	scope.App = &entity.App{ID: "app", ProjectID: "p", Project: &entity.Project{ID: "p"},
		ProjectEnv: &entity.ProjectEnv{Name: "dev"}}
	data := &queue.PeriodicExecData{
		PeriodicSetting: f.setting, Scope: scope, RefObjects: entity.NewRefObjects(),
		Task: &entity.Task{StartedAt: timeutil.NowUTC()},
	}
	_, err := f.svc.Healthcheck(context.Background(), &healthcheckservice.HealthcheckReq{
		PeriodicExecData: data, Healthcheck: f.setting.MustAsPeriodicJob().Healthcheck,
	})
	assert.NoError(t, err)
	return data.Task.Status, data.SaveTask
}

// A check tells a change of its state: failing, and healthy again after
// failing. Healthy from its first run it has nothing to say, and failing on
// it is not told again. The state kept is the latest result, which the next
// run tells a change by - it used to stay the first result ever, so every run
// after the first change read as one more.
func TestACheckTellsItsChangesOfState(t *testing.T) {
	f := newCheckFixture(t, 10*time.Second, 0)

	status, save := f.run(t)
	assert.Equal(t, base.TaskStatusDone, status)
	assert.True(t, save, "the first result is kept")
	assert.Empty(t, f.notifier.told, "healthy from the first run")

	f.status.Store(http.StatusInternalServerError)
	status, save = f.run(t)
	assert.Equal(t, base.TaskStatusFailed, status)
	assert.True(t, save, "a change")
	assert.Equal(t, []bool{false}, f.notifier.told, "failing is told")
	assert.Equal(t, []base.SchedJobTriggerEvent{base.SchedJobTriggerHealthDown}, f.triggers.fired)

	_, save = f.run(t)
	assert.False(t, save, "failing on is no change")
	assert.Equal(t, []bool{false}, f.notifier.told, "not told again")

	f.status.Store(http.StatusOK)
	_, save = f.run(t)
	assert.True(t, save, "a change")
	assert.Equal(t, []bool{false, true}, f.notifier.told, "healthy again is told")
	assert.Equal(t, []base.SchedJobTriggerEvent{base.SchedJobTriggerHealthDown, base.SchedJobTriggerHealthUp},
		f.triggers.fired)

	_, save = f.run(t)
	assert.False(t, save, "healthy on")
	assert.Len(t, f.notifier.told, 2)
}

// Failing on, a check is told again every repeat set on it - never while it
// is healthy on, whatever the repeat.
func TestAFailingCheckIsToldAgainEveryRepeat(t *testing.T) {
	now := time.Now()
	failingSince := &cacheentity.HealthcheckState{State: base.HealthcheckStateFailure,
		LastNotifTs: now.Add(-2 * time.Minute)}

	assert.False(t, shouldNotify(true, failingSince, 0, now), "no repeat set")
	assert.True(t, shouldNotify(true, failingSince, time.Minute, now), "a repeat due")
	assert.False(t, shouldNotify(true, failingSince, 5*time.Minute, now), "a repeat not yet due")
	assert.True(t, shouldNotify(true, &cacheentity.HealthcheckState{State: base.HealthcheckStateFailure},
		5*time.Minute, now), "failing, never told")

	healthySince := &cacheentity.HealthcheckState{State: base.HealthcheckStateSuccess,
		LastNotifTs: now.Add(-2 * time.Minute)}
	assert.False(t, shouldNotify(false, healthySince, time.Minute, now), "healthy on")

	assert.True(t, shouldNotify(true, nil, 0, now), "failing from the first run")
	assert.False(t, shouldNotify(false, nil, 0, now), "healthy from the first run")
}

// A check that sets no timeout is still bounded, by its interval at most: the
// client has no timeout of its own, and a server that never answers used to
// hold the run - and its slot among the runs - for the run's whole ceiling.
func TestACheckWithoutATimeoutGivesUpWithinItsInterval(t *testing.T) {
	f := newCheckFixture(t, time.Second, 0)
	f.hang.Store(true)

	started := time.Now()
	status, _ := f.run(t)

	assert.Equal(t, base.TaskStatusFailed, status)
	assert.Less(t, time.Since(started), 3*time.Second)
}

func TestAttemptTimeout(t *testing.T) {
	job := func(timeout, interval time.Duration) *entity.PeriodicJob {
		return &entity.PeriodicJob{Timeout: timeutil.Duration(timeout), Interval: timeutil.Duration(interval)}
	}
	assert.Equal(t, 5*time.Second, attemptTimeout(job(5*time.Second, time.Second)), "its own, as set")
	assert.Equal(t, 10*time.Second, attemptTimeout(job(0, 10*time.Second)), "none: its interval")
	assert.Equal(t, defaultAttemptTimeout, attemptTimeout(job(0, time.Hour)), "none: the default, within a long interval")
	assert.Equal(t, defaultAttemptTimeout, attemptTimeout(job(0, 0)))
}
