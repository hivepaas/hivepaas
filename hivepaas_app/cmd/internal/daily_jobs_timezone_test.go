package internal

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settinginitservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// txDB runs a transaction's function, and says nothing of committing.
type txDB struct {
	database.IDB
}

func (txDB) RunInTx(ctx context.Context, _ *sql.TxOptions, fn func(context.Context, bun.Tx) error) error {
	return fn(ctx, bun.Tx{})
}

type statusRepo struct {
	repository.SystemStatusRepo
	status *entity.SystemStatus
	saved  []string
}

func (r *statusRepo) Get(context.Context, database.IDB, ...bunex.SelectQueryOption) (*entity.SystemStatus, error) {
	return r.status, nil
}

func (r *statusRepo) Upsert(_ context.Context, _ database.IDB, status *entity.SystemStatus, _, cols []string,
	_ ...bunex.InsertQueryOption) error {
	r.status, r.saved = status, cols
	return nil
}

type movingJobs struct {
	settinginitservice.Service
	moved    []*entity.Setting
	from, to string
	calls    int
}

func (m *movingJobs) MoveDailyJobs(_ context.Context, _ database.Tx, from, to *time.Location) (
	[]*entity.Setting, []string, error) {
	m.calls++
	m.from, m.to = from.String(), to.String()
	return m.moved, []string{"System backup"}, nil
}

type rescheduling struct {
	queue.TaskQueue
	jobs       []*entity.Setting
	unschedule bool
}

func (q *rescheduling) ScheduleTasksForSchedJobs(_ context.Context, _ database.Tx, jobs []*entity.Setting,
	unschedule bool) error {
	q.jobs, q.unschedule = jobs, unschedule
	return nil
}

// The jobs are moved when the timezone is not the one they were given their
// times in - UTC, for an installation from before there was one - and their
// tasks made again; the status then says the new one, and a start under it
// moves nothing.
func TestDailyJobsMoveToTheTimezone(t *testing.T) {
	vn, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	assert.NoError(t, err)
	status := &statusRepo{status: &entity.SystemStatus{ID: "1"}}
	jobs := &movingJobs{moved: []*entity.Setting{{ID: "cleanup-job"}}}
	tasks := &rescheduling{}
	run := func(to *time.Location) error {
		return moveDailyJobsToTimezone(context.Background(), txDB{}, status, jobs, tasks, to, logging.GlobalLogger())
	}

	assert.NoError(t, run(vn))
	assert.Equal(t, 1, jobs.calls)
	assert.Equal(t, "UTC", jobs.from, "given before there was a timezone: in UTC")
	assert.Equal(t, "Asia/Ho_Chi_Minh", jobs.to)
	assert.Equal(t, []*entity.Setting{{ID: "cleanup-job"}}, tasks.jobs)
	assert.True(t, tasks.unschedule, "the tasks of the old times go")
	assert.Equal(t, "Asia/Ho_Chi_Minh", status.status.ScheduleTimezone)
	assert.Equal(t, entity.SystemStatusScheduleTimezoneCols, status.saved)

	assert.NoError(t, run(vn))
	assert.Equal(t, 1, jobs.calls, "already there")

	assert.NoError(t, run(time.UTC))
	assert.Equal(t, 2, jobs.calls)
	assert.Equal(t, "Asia/Ho_Chi_Minh", jobs.from)
	assert.Equal(t, "UTC", status.status.ScheduleTimezone)

	status.status.ScheduleTimezone = "Mars/Olympus_Mons"
	assert.Error(t, run(vn), "a timezone it cannot read is not taken for UTC")
}

// An installation from before there was a timezone, still in UTC, moves
// nothing.
func TestDailyJobsStayInUTC(t *testing.T) {
	status := &statusRepo{status: &entity.SystemStatus{ID: "1"}}
	jobs := &movingJobs{}
	assert.NoError(t, moveDailyJobsToTimezone(context.Background(), txDB{}, status, jobs, &rescheduling{},
		time.UTC, logging.GlobalLogger()))
	assert.Zero(t, jobs.calls)
	assert.Nil(t, status.saved)
}
