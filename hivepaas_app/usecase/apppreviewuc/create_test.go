package apppreviewuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apppreviewuc/apppreviewdto"
)

// previewPlan is the plan check, and what it was asked.
type previewPlan struct {
	licenseservice.Service
	err error

	db  database.IDB
	req *licenseservice.CheckReq
}

func (f *previewPlan) Check(_ context.Context, db database.IDB, req *licenseservice.CheckReq) error {
	f.db, f.req = db, req
	return f.err
}

// previewableApp is an app with previews on.
type previewableApp struct {
	appservice.Service
}

func (previewableApp) LoadAppWithFeatureSettings(_ context.Context, _ database.IDB, _, appID string, _, _ bool,
	_ ...bunex.SelectQueryOption) (*entity.App, *entity.AppFeatureSettings, error) {
	return &entity.App{ID: appID, Name: "web"},
		&entity.AppFeatureSettings{PreviewSettings: &entity.AppFeaturePreviewSettings{Enabled: true}}, nil
}

// previewTasks has no preview yet, and makes the preview's task; checked says
// whether the plan was checked by then.
type previewTasks struct {
	apppreviewservice.Service
	plan *previewPlan

	made    bool
	checked bool
}

func (previewTasks) GetPreview(context.Context, database.IDB, string, string, ...bunex.SelectQueryOption) (
	*entity.App, error) {
	return nil, hperrors.Wrap(hperrors.ErrNotFound)
}

func (f *previewTasks) CreateAppPreviewTask(app *entity.App, args *entity.TaskAppPreviewArgs) (*entity.Task, error) {
	f.made, f.checked = true, f.plan.req != nil
	return &entity.Task{ID: "task-1", ObjectID: app.ID, Type: base.TaskTypeAppPreview}, nil
}

type storedTasks struct {
	repository.TaskRepo
}

func (storedTasks) Upsert(context.Context, database.IDB, *entity.Task, []string, []string,
	...bunex.InsertQueryOption) error {
	return nil
}

type auditLog struct {
	auditservice.Service
}

func (auditLog) Record(context.Context, database.IDB, *auditservice.Entry) error {
	return nil
}

func newPreviewUC(plan *previewPlan, tasks *previewTasks) *UC {
	return &UC{
		taskRepo:          storedTasks{},
		appPreviewService: tasks,
		auditService:      auditLog{},
		appService:        previewableApp{},
		licenseService:    plan,
	}
}

var previewAuth = &basedto.Auth{User: &basedto.User{User: &entity.User{ID: "user-1"}}}

func previewReq() *apppreviewdto.CreatePreviewReq {
	return &apppreviewdto.CreatePreviewReq{ProjectID: "proj-1", AppID: "app-1", RepoRef: "pull/42"}
}

// The request is checked in its transaction, for one preview, for an answer on
// screen; refused, no task is made.
func TestCreatePreviewTaskChecksThePlan(t *testing.T) {
	plan := &previewPlan{err: hperrors.Wrap(hperrors.ErrPlanLimitReached)}
	tasks := &previewTasks{plan: plan}

	task, err := newPreviewUC(plan, tasks).createPreviewTask(context.Background(), database.Tx{}, previewAuth,
		previewReq())

	assert.ErrorIs(t, err, hperrors.ErrPlanLimitReached)
	assert.Nil(t, task)
	assert.False(t, tasks.made)
	assert.IsType(t, database.Tx{}, plan.db)
	assert.Equal(t, licenseservice.Adding(base.LicenseLimitPreviews, 1), plan.req)
}

func TestCreatePreviewTaskWithinThePlan(t *testing.T) {
	plan := &previewPlan{}
	tasks := &previewTasks{plan: plan}

	task, err := newPreviewUC(plan, tasks).createPreviewTask(context.Background(), database.Tx{}, previewAuth,
		previewReq())

	assert.NoError(t, err)
	if assert.NotNil(t, task) {
		assert.Equal(t, "task-1", task.ID)
	}
	assert.True(t, tasks.checked, "the plan is checked before the task is made")
}
