package nodeuc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice/schedjobserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type jobSettings struct {
	repository.SettingRepo
	job *entity.Setting
}

func (f *jobSettings) GetSingle(context.Context, database.IDB, *entity.ObjectScope, base.SettingType, bool,
	...bunex.SelectQueryOption) (*entity.Setting, error) {
	if f.job == nil {
		return nil, hperrors.NewNotFound("Scheduled job")
	}
	return f.job, nil
}

type insertedTasks struct {
	repository.TaskRepo
	tasks []*entity.Task
}

func (f *insertedTasks) Insert(_ context.Context, _ database.IDB, task *entity.Task,
	_ ...bunex.InsertQueryOption) error {
	f.tasks = append(f.tasks, task)
	return nil
}

type scheduledTasks struct {
	queue.TaskQueue
	tasks []*entity.Task
}

func (f *scheduledTasks) ScheduleTask(_ context.Context, tasks ...*entity.Task) error {
	f.tasks = append(f.tasks, tasks...)
	return nil
}

func artifactsJob() *entity.Setting {
	job := &entity.Setting{ID: "job-artifacts", Scope: base.ObjectScopeGlobal, Type: base.SettingTypeSchedJob,
		Kind: string(base.SchedJobTypeManagedArtifacts), Status: base.SettingStatusActive}
	job.MustSetData(&entity.SchedJob{JobType: base.SchedJobTypeManagedArtifacts,
		Schedule: &entity.SchedJobSchedule{Interval: 600}})
	return job
}

// A node joined has the managed artifacts' job run two minutes later, once it
// is Ready; with no job, nothing is.
func TestAJoinBringsTheArtifactsRunForward(t *testing.T) {
	now := time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)
	jobs, tasks, queued := &jobSettings{job: artifactsJob()}, &insertedTasks{}, &scheduledTasks{}
	uc := &UC{BaseUC: &settings.BaseUC{SettingRepo: jobs}, taskRepo: tasks, taskQueue: queued,
		schedJobService: schedjobserviceimpl.New(nil)}

	task, err := uc.scheduleArtifactsRun(context.Background(), now)

	assert.NoError(t, err)
	if assert.NotNil(t, task) {
		assert.Equal(t, "job-artifacts", task.TargetID)
		assert.Equal(t, now.Add(2*time.Minute), task.RunAt)
		assert.Equal(t, []*entity.Task{task}, tasks.tasks)
		assert.Equal(t, []*entity.Task{task}, queued.tasks)
	}

	jobs.job = nil
	task, err = uc.scheduleArtifactsRun(context.Background(), now)
	assert.NoError(t, err)
	assert.Nil(t, task)
}
