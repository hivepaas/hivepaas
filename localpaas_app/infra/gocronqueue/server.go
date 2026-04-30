package gocronqueue

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/pkg/redishelper"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	defaultConcurrency        = 5
	taskHighPriorityLookAhead = 1 * time.Second
	taskLowPriorityDelay      = 500 * time.Millisecond
)

var (
	ErrTaskProcessorNotFound = errors.New("task processor not found")
)

type TaskExecFunc func(taskID string, payload string) *time.Time

type Server struct {
	config    *Config
	scheduler gocron.Scheduler
	jobMap    map[string]*jobData // task.ID -> job data
	mu        sync.RWMutex
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

	// Healthcheck: a special kind of task
	HealthcheckBaseInterval time.Duration
	HealthcheckFunc         func(ctx context.Context) error
}

type jobData struct {
	Job      gocron.Job
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
		return nil, apperrors.Wrap(err)
	}
	return &Server{
		scheduler: scheduler,
		config:    config,
		jobMap:    make(map[string]*jobData, 20), //nolint:mnd
	}, nil
}

func (s *Server) Start() error {
	s.scheduler.Start()

	// Start a job to periodically check scheduling messages in redis
	go func() {
		for {
			s.listenToSchedMessages(context.Background())
		}
	}()

	// Start a job to periodically create new tasks from cron jobs
	go func() {
		for range time.Tick(s.config.TaskCreateInterval) {
			s.createTasks()
		}
	}()
	s.createTasks()

	// Start a job to periodically scan for new tasks from DB
	go func() {
		for range time.Tick(s.config.TaskCheckInterval) {
			s.scanTasks()
		}
	}()
	s.scanTasks()

	// Start a job to periodically do health check
	go func() {
		interval := s.config.HealthcheckBaseInterval
		timeNow := time.Now()
		wait := timeNow.Truncate(interval).Add(interval).Sub(timeNow)
		time.Sleep(wait)
		_, err := s.scheduler.NewJob(
			gocron.DurationJob(interval),
			gocron.NewTask(func() {
				err := s.config.HealthcheckFunc(context.Background())
				if err != nil {
					s.config.Logger.Errorf("failed to execute healthcheck task: %v", err)
				}
			}),
		)
		if err != nil {
			s.config.Logger.Errorf("failed to schedule healthcheck task: %v", err)
		}
	}()

	return nil
}

func (s *Server) listenToSchedMessages(ctx context.Context) {
	defer func() {
		_ = recover()
	}()

	// TODO: use BLMOVE to handle the case we fail to process the msg?
	schedMsg, err := redishelper.BLPopOne[*SchedMessage](ctx, s.config.RedisClient,
		taskQueueSchedKey, taskQueueSchedReadTimeout)
	if err != nil {
		time.Sleep(10 * time.Second) //nolint:mnd
		return
	}
	if len(schedMsg.SchedTasks) > 0 {
		err := s.ScheduleTask(ctx, schedMsg.SchedTasks...)
		if err != nil {
			s.config.Logger.Errorf("failed to schedule tasks from redis message: %v", err)
		}
		return
	}
	if len(schedMsg.UnschedTaskIDs) > 0 {
		err := s.UnscheduleTask(ctx, schedMsg.UnschedTaskIDs...)
		if err != nil {
			s.config.Logger.Errorf("failed to unschedule tasks from redis message: %v", err)
		}
		return
	}
}

func (s *Server) createTasks() {
	defer func() {
		if r := recover(); r != nil {
			s.config.Logger.Errorf("panic when create new tasks: %v", r)
		}
	}()
	err := s.config.TaskCreateFunc(context.Background())
	if err != nil {
		s.config.Logger.Errorf("failed to create new tasks: %v", err)
	}
}

func (s *Server) scanTasks() {
	defer func() {
		if r := recover(); r != nil {
			s.config.Logger.Errorf("panic when scan tasks for running: %v", r)
		}
	}()

	tasks, err := s.config.TaskCheckFunc(context.Background())
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
			return apperrors.Wrap(err)
		}
	}
	return nil
}

func (s *Server) scheduleTask(task *entity.Task, runAt time.Time) error {
	if !s.shouldSchedule(task, runAt) {
		return nil
	}
	var startAt gocron.OneTimeJobStartAtOption
	if runAt.IsZero() || runAt.Before(timeutil.NowUTC()) {
		startAt = gocron.OneTimeJobStartImmediately()
	} else {
		startAt = gocron.OneTimeJobStartDateTime(runAt)
	}
	job, err := s.scheduler.NewJob(
		gocron.OneTimeJob(startAt),
		gocron.NewTask(func() {
			err := s.executeTask(task, true)
			if err != nil {
				s.config.Logger.Errorf("failed to execute task '%v', id %s: %v", task.Type, task.ID, err)
			}
		}),
	)
	if err != nil {
		s.config.Logger.Errorf("failed to schedule task %s: %v", task.ID, err)
		return apperrors.Wrap(err)
	}
	s.addJob(task, job, runAt)
	return nil
}

func (s *Server) UnscheduleTask(ctx context.Context, taskIDs ...string) error {
	for _, taskID := range taskIDs {
		s.removeJob(taskID, true)
	}
	return nil
}

func (s *Server) shouldSchedule(task *entity.Task, runAt time.Time) bool {
	if s.config.TaskCanScheduleFunc != nil && !s.config.TaskCanScheduleFunc(task) {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	existingJob := s.jobMap[task.ID]
	if existingJob != nil && existingJob.RunAt.Equal(runAt) { // NOTE: zero time values equal
		return false
	}
	return true
}

func (s *Server) executeTask(task *entity.Task, priorityCheck bool) error {
	// Skip this task and queue it for running later if there is higher priority task
	if priorityCheck && task.Config.Priority != base.TaskPriorityCritical {
		priorityJob := s.findPriorityJob(task, timeutil.NowUTC())
		if priorityJob != nil {
			err := s.scheduleTask(task, priorityJob.RunAt.Add(taskLowPriorityDelay))
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		}
	}

	defer func() {
		s.removeJob(task.ID, false)
	}()

	execFunc := s.config.TaskMap[task.Type]
	if execFunc == nil {
		return fmt.Errorf("%w: task processor func not found for task type '%v'",
			ErrTaskProcessorNotFound, task.Type)
	}
	rescheduleAt := execFunc(task.ID, task.Args)
	if rescheduleAt != nil {
		err := s.scheduleTask(task, *rescheduleAt)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}

func (s *Server) addJob(task *entity.Task, job gocron.Job, runAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if currJob := s.jobMap[task.ID]; currJob != nil {
		if currJob.Job == job {
			return
		}
		_ = s.scheduler.RemoveJob(currJob.Job.ID())
	}
	s.jobMap[task.ID] = &jobData{
		Job:      job,
		RunAt:    runAt,
		Priority: task.Config.Priority,
	}
}

func (s *Server) removeJob(taskID string, unschedule bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if currJob := s.jobMap[taskID]; currJob != nil {
		if unschedule {
			_ = s.scheduler.RemoveJob(currJob.Job.ID())
		}
		delete(s.jobMap, taskID)
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
	return s.executeTask(task, false)
}

func (s *Server) Shutdown() error {
	if s.scheduler != nil {
		err := s.scheduler.Shutdown()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}
