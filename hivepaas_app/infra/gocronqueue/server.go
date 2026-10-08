package gocronqueue

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/redishelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	defaultConcurrency        = 10
	defaultPeriodicInterval   = 1 * time.Second
	taskHighPriorityLookAhead = 1 * time.Second
	taskLowPriorityDelay      = 500 * time.Millisecond

	// defaultPauseLimit is the longest a stop holds the scheduler. A stop comes
	// before a restart that replaces the process within minutes; one still here
	// after this was not replaced - the stop reached a process the restart left
	// alone, or the restart failed - and the pause ends on its own.
	defaultPauseLimit = 10 * time.Minute
	// defaultPauseRecheck is how often a task held by a pause looks again, and so
	// how late it runs once a start ends the pause.
	defaultPauseRecheck = 5 * time.Second
)

var (
	ErrTaskExecutorNotFound = errors.New("task executor not found")
)

type TaskExecFunc func(taskID string, payload string) (reschedAt time.Time)

type Server struct {
	config     *Config
	scheduler  gocron.Scheduler
	jobMap     map[string]*jobData // task.ID -> job data
	mu         sync.RWMutex
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
	// startedAt is when this process started its scheduler: a stop sent
	// before then was meant for another.
	startedAt time.Time

	// A stop pauses the scheduler rather than stopping gocron. Stopped, gocron
	// held every task for good when the restart the stop was for never came,
	// and dropped the one-time jobs whose time passed meanwhile once started
	// again. A task due while paused is held, looking again every
	// pauseRecheckEvery, until a start ends the pause or pauseLimit runs out.
	pauseLimit        time.Duration
	pauseRecheckEvery time.Duration
	pauseMu           sync.Mutex
	pausedUntil       time.Time
}

type Config struct {
	Concurrency int
	TaskMap     map[base.TaskType]TaskExecFunc
	RedisClient redis.UniversalClient
	Logger      logging.Logger

	TaskCheckFunc       func(ctx context.Context) ([]*entity.Task, error)
	TaskCheckInterval   time.Duration
	TaskCreateFunc      func(ctx context.Context) error
	TaskCreateInterval  time.Duration
	TaskCanScheduleFunc func(*entity.Task) bool

	// Periodic: a special kind of task
	PeriodicBaseInterval time.Duration
	PeriodicExecFunc     func(ctx context.Context) error
}

type jobData struct {
	Job      gocron.Job // nil until the scheduler has made it
	RunAt    time.Time
	Priority base.TaskPriority
}

func NewServer(config *Config) (*Server, error) {
	if config.Concurrency <= 0 {
		config.Concurrency = defaultConcurrency
	}
	scheduler, err := gocron.NewScheduler(
		gocron.WithLimitConcurrentJobs(uint(config.Concurrency), gocron.LimitModeWait),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &Server{
		scheduler:         scheduler,
		config:            config,
		jobMap:            make(map[string]*jobData, 20), //nolint:mnd
		pauseLimit:        defaultPauseLimit,
		pauseRecheckEvery: defaultPauseRecheck,
	}, nil
}

func (s *Server) Start() error {
	s.startedAt = timeutil.NowUTC()
	s.scheduler.Start()

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	// Start a job to periodically check controlling messages in redis
	s.wg.Go(func() {
		defer safego.RecoverWithLogger(s.config.Logger, "gocronqueue.listenToCtrlMessages")
		for {
			if ctx.Err() != nil {
				return
			}
			s.listenToCtrlMessages(ctx)
		}
	})

	// Start a job to periodically create new tasks from cron jobs
	s.wg.Go(func() {
		defer safego.RecoverWithLogger(s.config.Logger, "gocronqueue.createTasks")
		ticker := time.NewTicker(s.config.TaskCreateInterval)
		defer ticker.Stop()
		s.createTasks(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.createTasks(ctx)
			}
		}
	})

	// Start a job to periodically scan for new tasks from DB
	s.wg.Go(func() {
		defer safego.RecoverWithLogger(s.config.Logger, "gocronqueue.scanTasks")
		ticker := time.NewTicker(s.config.TaskCheckInterval)
		defer ticker.Stop()
		s.scanTasks(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.scanTasks(ctx)
			}
		}
	})

	// Start a job to execute periodic tasks
	s.wg.Go(func() {
		defer safego.RecoverWithLogger(s.config.Logger, "gocronqueue.execPeriodicJob")
		interval := s.config.PeriodicBaseInterval
		if interval <= 0 {
			interval = defaultPeriodicInterval
		}
		timeNow := time.Now()
		wait := timeNow.Truncate(interval).Add(interval).Sub(timeNow)

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}

		s.execPeriodicJob(ctx)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.execPeriodicJob(ctx)
			}
		}
	})

	return nil
}

// execPeriodicJob runs one round of the periodic job. A panic here used to kill
// the whole process, unlike createTasks/scanTasks which were already guarded.
func (s *Server) execPeriodicJob(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			s.config.Logger.Errorf("panic when executing periodic job: %v\n%s", r, debug.Stack())
		}
	}()
	if err := s.config.PeriodicExecFunc(ctx); err != nil {
		s.config.Logger.Errorf("failed to execute periodic job: %v", err)
	}
}

func (s *Server) listenToCtrlMessages(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			s.config.Logger.Errorf("panic when handling control messages: %v\n%s", r, debug.Stack())
		}
	}()

	// TODO: use BLMOVE to handle the case we fail to process the msg?
	ctrlMsg, err := redishelper.BLPopOne[*Message](ctx, s.config.RedisClient,
		taskQueueCtrlKey, taskQueueCtrlReadTimeout)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		if wait := ctrlReadBackoff(err); wait > 0 {
			select {
			case <-ctx.Done():
			case <-time.After(wait):
			}
		}
		return
	}
	s.handleCtrlMessage(ctx, ctrlMsg)
}

// ctrlReadBackoff is how long to wait before reading the list again: not at all
// when it was only empty for the read's timeout, which is every few seconds.
func ctrlReadBackoff(err error) time.Duration {
	if errors.Is(err, hperrors.ErrNotFound) {
		return 0
	}
	return ctrlReadErrorBackoff
}

// handleCtrlMessage does what a message asks of this process.
func (s *Server) handleCtrlMessage(ctx context.Context, ctrlMsg *Message) {
	if ctrlMsg == nil {
		return
	}
	if ctrlMsg.StartScheduler {
		s.resume(fmt.Sprintf("a control message sent at %v", ctrlMsg.SentAt))
		return
	}
	if ctrlMsg.StopScheduler {
		// A stop is for the processes running when it was sent - an update's,
		// for those it is about to replace. One that reaches a process started
		// since is left from a restart already done, and would hold its tasks
		// for nothing: nothing sends a start after an update.
		if ctrlMsg.SentAt.Before(s.startedAt) {
			s.config.Logger.Warnf("task queue scheduler: ignored a stop sent at %v, before this process "+
				"started at %v", ctrlMsg.SentAt, s.startedAt)
			return
		}
		s.pause(fmt.Sprintf("a control message sent at %v", ctrlMsg.SentAt))
		return
	}

	if len(ctrlMsg.SchedTasks) > 0 {
		err := s.ScheduleTask(ctx, ctrlMsg.SchedTasks...)
		if err != nil {
			s.config.Logger.Errorf("failed to schedule tasks from redis message: %v", err)
		}
		return
	}
	if len(ctrlMsg.UnschedTaskIDs) > 0 {
		err := s.UnscheduleTask(ctx, ctrlMsg.UnschedTaskIDs...)
		if err != nil {
			s.config.Logger.Errorf("failed to unschedule tasks from redis message: %v", err)
		}
		return
	}
}

func (s *Server) createTasks(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			s.config.Logger.Errorf("panic when create new tasks: %v\n%s", r, debug.Stack())
		}
	}()
	err := s.config.TaskCreateFunc(ctx)
	if err != nil {
		s.config.Logger.Errorf("failed to create new tasks: %v", err)
	}
}

func (s *Server) scanTasks(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			s.config.Logger.Errorf("panic when scan tasks for running: %v\n%s", r, debug.Stack())
		}
	}()

	tasks, err := s.config.TaskCheckFunc(ctx)
	if err != nil {
		s.config.Logger.Errorf("failed to scan new tasks: %v", err)
		return
	}
	for _, task := range tasks {
		err = s.scheduleTask(task, task.ShouldRunAt())
		if err != nil {
			s.config.Logger.Errorf("failed to schedule new tasks: %v", err)
			return
		}
	}
}

func (s *Server) ScheduleTask(ctx context.Context, tasks ...*entity.Task) error {
	for _, task := range tasks {
		err := s.scheduleTask(task, task.ShouldRunAt())
		if err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (s *Server) scheduleTask(task *entity.Task, runAt time.Time) error {
	if runAt.IsZero() {
		return nil
	}
	if s.config.TaskCanScheduleFunc != nil && !s.config.TaskCanScheduleFunc(task) {
		return nil
	}
	data := &jobData{RunAt: runAt, Priority: task.Config.Priority}
	if !s.reserveJob(task.ID, data) {
		return nil
	}
	var startAt gocron.OneTimeJobStartAtOption
	if !runAt.After(timeutil.NowUTC()) {
		startAt = gocron.OneTimeJobStartImmediately()
	} else {
		startAt = gocron.OneTimeJobStartDateTime(runAt)
	}
	job, err := s.scheduler.NewJob(
		gocron.OneTimeJob(startAt),
		gocron.NewTask(func() {
			err := s.executeTask(task, data, true)
			if err != nil {
				s.config.Logger.Errorf("failed to execute task '%v', id %s: %v", task.Type, task.ID, err)
			}
		}),
	)
	if err != nil {
		s.releaseJob(task.ID, data)
		s.config.Logger.Errorf("failed to schedule task %s: %v", task.ID, err)
		return hperrors.Wrap(err)
	}
	s.setJob(task.ID, data, job)
	return nil
}

func (s *Server) UnscheduleTask(ctx context.Context, taskIDs ...string) error {
	for _, taskID := range taskIDs {
		s.removeJob(taskID)
	}
	return nil
}

// reserveJob makes data the task's entry, unless the task is scheduled for that
// time already, before data has a job: one due at once can run before the
// scheduler hands it back, and what its run puts in the map - a hold, a
// reschedule - is newer than data.
func (s *Server) reserveJob(taskID string, data *jobData) bool {
	s.mu.Lock()
	currJob := s.jobMap[taskID]
	if currJob != nil && currJob.RunAt.Equal(data.RunAt) {
		s.mu.Unlock()
		return false
	}
	s.jobMap[taskID] = data
	var replaced gocron.Job
	if currJob != nil {
		replaced = currJob.Job
	}
	s.mu.Unlock()
	if replaced != nil {
		_ = s.scheduler.RemoveJob(replaced.ID())
	}
	return true
}

// setJob gives data its job. A job whose entry was replaced meanwhile is removed
// from the scheduler: it has run already, or a newer one stands for the task.
func (s *Server) setJob(taskID string, data *jobData, job gocron.Job) {
	s.mu.Lock()
	data.Job = job
	replaced := s.jobMap[taskID] != data
	s.mu.Unlock()
	if replaced {
		_ = s.scheduler.RemoveJob(job.ID())
	}
}

// releaseJob removes data from the map, if it is still the task's entry.
func (s *Server) releaseJob(taskID string, data *jobData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.jobMap[taskID] == data {
		delete(s.jobMap, taskID)
	}
}

// executeTask runs the task of the entry own; own is nil when no job runs it.
func (s *Server) executeTask(task *entity.Task, own *jobData, priorityCheck bool) error {
	// Held while paused: it looks again shortly.
	now := timeutil.NowUTC()
	if s.paused(now) {
		err := s.scheduleTask(task, now.Add(s.pauseRecheckEvery))
		if err != nil {
			return hperrors.Wrap(err)
		}
		return nil
	}

	// Skip this task and queue it for running later if there is higher priority task
	if priorityCheck && task.Config.Priority != base.TaskPriorityCritical {
		priorityJob := s.findPriorityJob(task, now)
		if priorityJob != nil {
			err := s.scheduleTask(task, priorityJob.RunAt.Add(taskLowPriorityDelay))
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		}
	}

	var rescheduled bool
	defer func() {
		if !rescheduled && own != nil {
			s.releaseJob(task.ID, own)
		}
	}()

	execFunc := s.config.TaskMap[task.Type]
	if execFunc == nil {
		return fmt.Errorf("%w: task executor func not found for task type '%v'",
			ErrTaskExecutorNotFound, task.Type)
	}
	rescheduleAt := execFunc(task.ID, task.Args)
	if !rescheduleAt.IsZero() {
		err := s.scheduleTask(task, rescheduleAt)
		if err == nil {
			rescheduled = true
		}
		if err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

// removeJob unschedules the task. A job still being made is removed by setJob.
func (s *Server) removeJob(taskID string) {
	s.mu.Lock()
	var job gocron.Job
	if currJob := s.jobMap[taskID]; currJob != nil {
		job = currJob.Job
		delete(s.jobMap, taskID)
	}
	s.mu.Unlock()
	if job != nil {
		_ = s.scheduler.RemoveJob(job.ID())
	}
}

func (s *Server) findPriorityJob(currentTask *entity.Task, runAt time.Time) *jobData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for taskID, job := range s.jobMap {
		if taskID == currentTask.ID {
			continue
		}
		if job.Priority.Cmp(currentTask.Config.Priority) <= 0 {
			continue
		}
		diff := job.RunAt.Sub(runAt)
		if -taskHighPriorityLookAhead < diff && diff < taskHighPriorityLookAhead {
			return job
		}
	}
	return nil
}

func (s *Server) ScheduleNextTask(task *entity.Task, _ time.Time) error {
	return s.executeTask(task, nil, false)
}

func (s *Server) Shutdown() error {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
	s.wg.Wait()

	if s.scheduler != nil {
		err := s.scheduler.Shutdown()
		if err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

// StartScheduler ends a pause: the tasks it held run within pauseRecheckEvery.
func (s *Server) StartScheduler() error {
	s.resume("this process")
	return nil
}

// StopScheduler pauses the scheduler, for a restart about to replace this process.
func (s *Server) StopScheduler() error {
	s.pause("this process")
	return nil
}

func (s *Server) pause(asker string) {
	s.pauseMu.Lock()
	s.pausedUntil = timeutil.NowUTC().Add(s.pauseLimit)
	s.pauseMu.Unlock()
	s.config.Logger.Infof("task queue scheduler paused for at most %v, as %s asked", s.pauseLimit, asker)
}

func (s *Server) resume(asker string) {
	s.pauseMu.Lock()
	wasPaused := !s.pausedUntil.IsZero()
	s.pausedUntil = time.Time{}
	s.pauseMu.Unlock()
	if wasPaused {
		s.config.Logger.Infof("task queue scheduler resumed, as %s asked", asker)
	}
}

// paused says whether the scheduler is paused at now. A pause that has run out
// ends here, and says so: the restart it was for did not replace this process.
func (s *Server) paused(now time.Time) bool {
	s.pauseMu.Lock()
	until := s.pausedUntil
	ranOut := !until.IsZero() && !now.Before(until)
	if ranOut {
		s.pausedUntil = time.Time{}
	}
	s.pauseMu.Unlock()
	if ranOut {
		s.config.Logger.Warnf("task queue scheduler resumed on its own: its pause ran out at %v, "+
			"and no restart had replaced this process", until)
	}
	return !until.IsZero() && !ranOut
}
