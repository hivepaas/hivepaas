package internal

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type certRequestingService struct {
	getstartedservice.Service
	tasks            []*entity.Task
	notAsked         string
	err              error
	ignoreRetryAfter bool
}

func (s *certRequestingService) RequestDashboardCert(
	_ context.Context, _ database.IDB, ignoreRetryAfter bool,
) (*getstartedservice.CertRequest, error) {
	s.ignoreRetryAfter = ignoreRetryAfter
	if s.err != nil {
		return nil, s.err
	}
	return &getstartedservice.CertRequest{Tasks: s.tasks, NotAsked: s.notAsked}, nil
}

type schedulingQueue struct {
	queue.TaskQueue
	scheduled []*entity.Task
	err       error
}

func (q *schedulingQueue) ScheduleTask(_ context.Context, tasks ...*entity.Task) error {
	q.scheduled = append(q.scheduled, tasks...)
	return q.err
}

type errorLogger struct {
	logging.Logger
	errors   []string
	warnings []string
}

func (l *errorLogger) Errorf(template string, args ...any) {
	l.errors = append(l.errors, fmt.Sprintf(template, args...))
}

func (l *errorLogger) Warnf(template string, args ...any) {
	l.warnings = append(l.warnings, fmt.Sprintf(template, args...))
}

func TestFirstBootSchedulesTheDashboardCertificate(t *testing.T) {
	task := &entity.Task{ID: "task-1"}
	service := &certRequestingService{tasks: []*entity.Task{task}}
	taskQueue := &schedulingQueue{}
	logger := &errorLogger{}

	requestDashboardCert(context.Background(), nil, service, taskQueue, logger)

	assert.False(t, service.ignoreRetryAfter, "the first boot keeps the wait a failure leaves")
	assert.Equal(t, []*entity.Task{task}, taskQueue.scheduled)
	assert.Empty(t, logger.errors)
}

func TestFirstBootGoesOnWhenTheCertificateCannotBeAskedFor(t *testing.T) {
	service := &certRequestingService{err: errors.New("no routing")}
	taskQueue := &schedulingQueue{}
	logger := &errorLogger{}

	requestDashboardCert(context.Background(), nil, service, taskQueue, logger)

	assert.Empty(t, taskQueue.scheduled)
	if assert.Len(t, logger.errors, 1) {
		assert.Contains(t, logger.errors[0], "no routing")
	}
}

func TestFirstBootLogsATaskItCannotSchedule(t *testing.T) {
	service := &certRequestingService{tasks: []*entity.Task{{ID: "task-1"}}}
	taskQueue := &schedulingQueue{err: errors.New("queue down")}
	logger := &errorLogger{}

	requestDashboardCert(context.Background(), nil, service, taskQueue, logger)

	if assert.Len(t, logger.errors, 1) {
		assert.Contains(t, logger.errors[0], "queue down")
	}
}

func TestFirstBootSchedulesNothingWhenNothingIsAskedFor(t *testing.T) {
	taskQueue := &schedulingQueue{}
	logger := &errorLogger{}

	requestDashboardCert(context.Background(), nil, &certRequestingService{}, taskQueue, logger)

	assert.Empty(t, taskQueue.scheduled)
	assert.Empty(t, logger.errors)
}

func TestFirstBootSaysWhyItAskedForNothing(t *testing.T) {
	service := &certRequestingService{notAsked: "no certificate was asked for localhost: not a public name"}
	taskQueue := &schedulingQueue{}
	logger := &errorLogger{}

	requestDashboardCert(context.Background(), nil, service, taskQueue, logger)

	assert.Empty(t, taskQueue.scheduled)
	assert.Empty(t, logger.errors)
	if assert.Len(t, logger.warnings, 1) {
		assert.Contains(t, logger.warnings[0], "not a public name")
	}
}
