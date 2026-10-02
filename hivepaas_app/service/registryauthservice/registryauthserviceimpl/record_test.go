package registryauthserviceimpl

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobservice"
)

// renewalSettings has the renewal's setting and job, each on or off, and
// the ECR credentials that use key auth ka1.
type renewalSettings struct {
	settingStore
	renewalOn, jobOn bool
	byKeyAuth        []*entity.Setting
}

func (f *renewalSettings) GetSingle(_ context.Context, _ database.IDB, _ *entity.ObjectScope,
	typ base.SettingType, _ bool, _ ...bunex.SelectQueryOption) (*entity.Setting, error) {
	if typ == base.SettingTypeRegistryAuthRenewal && f.renewalOn {
		return &entity.Setting{ID: "renewal", Type: typ, Status: base.SettingStatusActive}, nil
	}
	if typ == base.SettingTypeSchedJob && f.jobOn {
		return &entity.Setting{ID: "job", Type: typ, Status: base.SettingStatusActive}, nil
	}
	return nil, hperrors.NewNotFound("Setting")
}

func (f *renewalSettings) List(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
	paging *basedto.Paging, opts ...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	var rows []*entity.Setting
	sql := bunex.ApplySelect(bun.NewDB(nil, pgdialect.New()).NewSelect().Model(&rows), opts...).String()
	if strings.Contains(sql, "'keyAuth'") && strings.Contains(sql, "'ka1'") {
		return f.byKeyAuth, nil, nil
	}
	return f.settingStore.List(ctx, db, scope, paging, opts...)
}

type fakeJobs struct{ schedjobservice.Service }

func (fakeJobs) CreateSchedJobTask(job *entity.Setting, runAt, _ time.Time) (*entity.Task, error) {
	return &entity.Task{ID: "task-" + job.ID, TargetID: job.ID, RunAt: runAt}, nil
}

type taskRows struct {
	repository.TaskRepo
	inserted []*entity.Task
}

func (f *taskRows) Insert(_ context.Context, _ database.IDB, task *entity.Task, _ ...bunex.InsertQueryOption) error {
	f.inserted = append(f.inserted, task)
	return nil
}

func newRecordService(settings repository.SettingRepo, tasks *taskRows) *service {
	svc := newTestService(nil, &fakeECR{})
	svc.settingRepo, svc.taskRepo, svc.schedJobService = settings, tasks, fakeJobs{}
	return svc
}

// A renewal is recorded for the credentials named, as a task of the renewal's
// job; none while the renewal or its job is off.
func TestRecordRenewal(t *testing.T) {
	for _, off := range []*renewalSettings{{renewalOn: false, jobOn: true}, {renewalOn: true, jobOn: false}} {
		tasks := &taskRows{}
		task, err := newRecordService(off, tasks).RecordRenewal(context.Background(), nil, []string{"ra1"})
		assert.NoError(t, err)
		assert.Nil(t, task)
		assert.Empty(t, tasks.inserted)
	}

	tasks := &taskRows{}
	task, err := newRecordService(&renewalSettings{renewalOn: true, jobOn: true}, tasks).
		RecordRenewal(context.Background(), nil, []string{"ra1", "ra2"})
	assert.NoError(t, err)
	if assert.NotNil(t, task) && assert.Len(t, tasks.inserted, 1) {
		assert.Equal(t, "job", task.TargetID)
		args, err := task.ArgsAsRegistryAuthRenewal()
		assert.NoError(t, err)
		assert.Equal(t, []string{"ra1", "ra2"}, args.TargetAuths.ToIDStringSlice())
	}
}

// For a key auth, the renewal is for the credentials using it; nothing is
// recorded when none do.
func TestRecordRenewalForKeyAuth(t *testing.T) {
	settings := &renewalSettings{renewalOn: true, jobOn: true,
		byKeyAuth: []*entity.Setting{{ID: "ra1"}, {ID: "ra3"}}}
	tasks := &taskRows{}
	task, err := newRecordService(settings, tasks).RecordRenewalForKeyAuth(context.Background(), nil, "ka1")
	assert.NoError(t, err)
	if assert.NotNil(t, task) {
		args, _ := task.ArgsAsRegistryAuthRenewal()
		assert.Equal(t, []string{"ra1", "ra3"}, args.TargetAuths.ToIDStringSlice())
	}

	tasks = &taskRows{}
	task, err = newRecordService(&renewalSettings{renewalOn: true, jobOn: true}, tasks).
		RecordRenewalForKeyAuth(context.Background(), nil, "ka-unused")
	assert.NoError(t, err)
	assert.Nil(t, task)
	assert.Empty(t, tasks.inserted)
}
