package appcloneserviceimpl

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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// fakeLicenseService records each plan check, and refuses with err.
type fakeLicenseService struct {
	licenseservice.Service
	dbs  []database.IDB
	reqs []*licenseservice.CheckReq
	err  error
}

func (f *fakeLicenseService) Check(_ context.Context, db database.IDB, req *licenseservice.CheckReq) error {
	f.dbs = append(f.dbs, db)
	f.reqs = append(f.reqs, req)
	return f.err
}

// fakeKindRepo lists the source app's kind, whatever it is asked.
type fakeKindRepo struct {
	repository.SettingRepo
	kinds []*entity.Setting
}

func (f *fakeKindRepo) List(
	context.Context, database.IDB, *entity.ObjectScope, *basedto.Paging, ...bunex.SelectQueryOption,
) ([]*entity.Setting, *basedto.PagingMeta, error) {
	return f.kinds, nil, nil
}

// fakeCloneNotifications sends nothing, and keeps what it was asked to send.
type fakeCloneNotifications struct {
	notificationservice.Service
	sent []*notificationservice.TaskResultNotificationReq
}

func (f *fakeCloneNotifications) GetNotificationForEvent(
	context.Context, database.IDB, *entity.ObjectScope, *entity.BaseEventNotification, bool, *entity.RefObjects,
) (*entity.Notification, error) {
	return &entity.Notification{}, nil
}

func (f *fakeCloneNotifications) BuildTitlePrefix(*entity.Project, *entity.App, *entity.User) string {
	return "[shop/blog]"
}

func (f *fakeCloneNotifications) NotifyForTaskResult(
	_ context.Context, _ database.IDB, req *notificationservice.TaskResultNotificationReq,
) (*notificationservice.TaskResultNotificationResp, error) {
	f.sent = append(f.sent, req)
	return &notificationservice.TaskResultNotificationResp{}, nil
}

func kindSetting(t *testing.T, category base.AppCategory) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: "set-kind", Type: base.SettingTypeAppKind, ObjectID: "src-1"}
	assert.NoError(t, setting.SetData(&entity.AppKindSettings{Category: category}))
	return setting
}

// A clone counts against what it will be: a function's clone that takes its
// code is a function, any other clone an app. It is checked on the clone task's
// transaction, which holds the plan-limit lock until the clone is committed.
func TestTheCloneTaskChecksWhatTheCloneWillBe(t *testing.T) {
	for name, tc := range map[string]struct {
		category   base.AppCategory
		withSource bool
		want       base.LicenseLimit
	}{
		"a function, with its code":    {base.AppCategoryFunction, true, base.LicenseLimitFunctions},
		"a function, without its code": {base.AppCategoryFunction, false, base.LicenseLimitApps},
		"a webapp":                     {base.AppCategoryWebapp, true, base.LicenseLimitApps},
	} {
		licenses := &fakeLicenseService{}
		svc := &service{licenseService: licenses, settingRepo: &fakeKindRepo{
			kinds: []*entity.Setting{kindSetting(t, tc.category)},
		}}
		data := &appCloneData{AppCloneReq: &appcloneservice.AppCloneReq{
			SrcApp:         &entity.App{ID: "src-1"},
			CloneSettings:  &entity.AppCloneSettings{CloneDeploymentSettings: tc.withSource},
			CheckPlanLimit: true,
		}}

		err := svc.checkPlanLimit(context.Background(), database.Tx{}, data)

		assert.NoError(t, err, name)
		if assert.Len(t, licenses.reqs, 1, name) {
			assert.IsType(t, database.Tx{}, licenses.dbs[0], name)
			assert.Equal(t, licenseservice.Adding(tc.want, 1), licenses.reqs[0], name)
		}
	}
}

// An app with no kind setting is a webapp, and its clone an app.
func TestTheCloneOfAnAppWithoutAKindIsAnApp(t *testing.T) {
	licenses := &fakeLicenseService{}
	svc := &service{licenseService: licenses, settingRepo: &fakeKindRepo{}}
	data := &appCloneData{AppCloneReq: &appcloneservice.AppCloneReq{
		SrcApp:         &entity.App{ID: "src-1"},
		CloneSettings:  &entity.AppCloneSettings{CloneDeploymentSettings: true},
		CheckPlanLimit: true,
	}}

	assert.NoError(t, svc.checkPlanLimit(context.Background(), database.Tx{}, data))
	assert.Equal(t, []*licenseservice.CheckReq{licenseservice.Adding(base.LicenseLimitApps, 1)}, licenses.reqs)
}

// A preview clones the app it previews, and a preview is counted as one: its
// clone is not checked against the apps.
func TestACloneNotAskedToIsNotChecked(t *testing.T) {
	licenses := &fakeLicenseService{err: hperrors.Wrap(hperrors.ErrPlanLimitReached)}
	svc := &service{licenseService: licenses, settingRepo: &fakeKindRepo{}}
	data := &appCloneData{AppCloneReq: &appcloneservice.AppCloneReq{
		SrcApp: &entity.App{ID: "src-1"}, CloneSettings: &entity.AppCloneSettings{},
	}}

	assert.NoError(t, svc.checkPlanLimit(context.Background(), database.Tx{}, data))
	assert.Empty(t, licenses.reqs)
}

// A clone refused by the plan fails before anything of it is made, and is told
// to the clone's notification targets like any other failed clone.
func TestARefusedCloneIsNotifiedLikeAnyFailedClone(t *testing.T) {
	onProd(t)
	src := srcAppWithCloneSettings(t, &entity.AppCloneSettings{
		TargetName: "blog-copy", Notification: &entity.BaseEventNotification{FailureUseDefault: true},
	})
	src.Project = &entity.Project{ID: "p1", Key: "shop", Name: "Shop"}
	src.ProjectEnvID = "p1:prod"
	src.ProjectEnv = &entity.ProjectEnv{ID: "p1:prod", ProjectID: "p1", Key: "prod", Name: "prod"}
	licenses := &fakeLicenseService{err: hperrors.Wrap(hperrors.ErrPlanLimitReached)}
	notifications := &fakeCloneNotifications{}
	svc := &service{
		appService:          &fakeCloneAppService{app: src},
		settingService:      &fakeCloneSettingService{},
		settingRepo:         &fakeKindRepo{},
		licenseService:      licenses,
		notificationService: notifications,
	}
	task := cloneTask(t)
	execData := &queue.TaskExecData{Task: task}

	_, err := svc.CloneApp(context.Background(), database.Tx{},
		&appcloneservice.AppCloneReq{TaskExecData: execData, CheckPlanLimit: true})

	assert.ErrorIs(t, err, hperrors.ErrPlanLimitReached)
	// What the queue does with a failed task: save it as failed, then run what
	// waits for the transaction's end.
	task.Status = base.TaskStatusFailed
	if assert.NotNil(t, execData.OnPostTxFunc, "the notification waits for the task's end") {
		execData.OnPostTxFunc()
	}
	if assert.Len(t, notifications.sent, 1) {
		assert.False(t, notifications.sent[0].ActionSucceeded)
		assert.Equal(t, notificationservice.TemplateAppCloneNotification, notifications.sent[0].TemplateName)
	}
}
