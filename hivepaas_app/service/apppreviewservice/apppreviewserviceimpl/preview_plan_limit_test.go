package apppreviewserviceimpl

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// planLicense is the license in force, and the checks asked of it.
type planLicense struct {
	licenseservice.Service
	state      *license.State
	currentErr error
	checkErr   error

	currentCalls int
	checkDB      database.IDB
	checkReq     *licenseservice.CheckReq
}

func (f *planLicense) Current(context.Context) (*license.State, error) {
	f.currentCalls++
	return f.state, f.currentErr
}

func (f *planLicense) Check(_ context.Context, db database.IDB, req *licenseservice.CheckReq) error {
	f.checkDB, f.checkReq = db, req
	return f.checkErr
}

func beeLicense() *planLicense {
	return &planLicense{state: &license.State{
		Status: base.LicenseStatusNone,
		Plan:   base.LicensePlanBee,
		Limits: license.Limits{base.LicenseLimitPreviews: 2},
	}}
}

// planNotifications resolves every target to one notification, and keeps what
// was sent.
type planNotifications struct {
	notificationservice.Service
	notification *entity.Notification

	eventSetting *entity.BaseEventNotification
	eventSuccess *bool
	sent         []*notificationservice.TaskResultNotificationReq
}

func (f *planNotifications) GetNotificationForEvent(_ context.Context, _ database.IDB, _ *entity.ObjectScope,
	eventSetting *entity.BaseEventNotification, eventSuccess bool, _ *entity.RefObjects) (
	*entity.Notification, error) {
	f.eventSetting, f.eventSuccess = eventSetting, &eventSuccess
	return f.notification, nil
}

func (f *planNotifications) NotifyForTaskResult(_ context.Context, _ database.IDB,
	req *notificationservice.TaskResultNotificationReq) (*notificationservice.TaskResultNotificationResp, error) {
	f.sent = append(f.sent, req)
	return &notificationservice.TaskResultNotificationResp{}, nil
}

func (f *planNotifications) BuildTitlePrefix(project *entity.Project, app *entity.App, _ *entity.User) string {
	return "[" + project.Name + "][" + app.Name + "]"
}

// previewParentApps loads the app a preview is asked of.
type previewParentApps struct {
	appservice.Service
	app *entity.App
}

func (f *previewParentApps) LoadApp(_ context.Context, _ database.IDB, _, _ string, _, _ bool,
	_ ...bunex.SelectQueryOption) (*entity.App, error) {
	return f.app, nil
}

// noPreviews has no preview of any app.
type noPreviews struct {
	repository.AppRepo
}

func (noPreviews) List(_ context.Context, _ database.IDB, _ string, _ *basedto.Paging,
	_ ...bunex.SelectQueryOption) ([]*entity.App, *basedto.PagingMeta, error) {
	return nil, nil, nil
}

var errCloneReached = errors.New("clone reached")

// cloneAfterCheck stops a preview at its clone, and says whether the plan was
// checked by then.
type cloneAfterCheck struct {
	appcloneservice.Service
	lic *planLicense

	called       bool
	checkedFirst bool
}

func (f *cloneAfterCheck) CloneApp(context.Context, database.IDB, *appcloneservice.AppCloneReq) (
	*appcloneservice.AppCloneResp, error) {
	f.called, f.checkedFirst = true, f.lic.checkReq != nil
	return nil, errCloneReached
}

var planNotification = &entity.BaseEventNotification{Failure: entity.ObjectID{ID: "notif-1"}}

// previewParentApp is an app that deploys from a repository, and tells
// planNotification when a deployment fails.
func previewParentApp(t *testing.T, notification *entity.BaseEventNotification) *entity.App {
	t.Helper()
	project := &entity.Project{ID: "proj-1", Name: "shop"}
	app := &entity.App{
		ID: "app-1", Name: "web", ProjectID: project.ID, Project: project,
		ProjectEnv: &entity.ProjectEnv{ID: "env-1", ProjectID: project.ID, Name: "prod"},
	}
	setting := &entity.Setting{ID: "set-1", Type: base.SettingTypeAppDeployment}
	setting.MustSetData(&entity.AppDeploymentSettings{
		ActiveMethod: base.DeploymentMethodRepo,
		RepoSource:   &entity.DeploymentRepoSource{RepoURL: "https://github.com/acme/web", RepoRef: "main"},
		Notification: notification,
	})
	app.Settings = []*entity.Setting{setting}
	return app
}

func previewTaskExecData(t *testing.T, repoRef string) *queue.TaskExecData {
	t.Helper()
	task := &entity.Task{ID: "task-1", Type: base.TaskTypeAppPreview,
		StartedAt: time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)}
	if err := task.SetArgs(&entity.TaskAppPreviewArgs{
		ParentApp: entity.ObjectID{ID: "app-1"},
		RepoRef:   repoRef,
	}); err != nil {
		t.Fatal(err)
	}
	return &queue.TaskExecData{Task: task}
}

func useDashboard(t *testing.T) {
	t.Helper()
	previous := config.Current()
	config.SetCurrent(&config.Config{})
	t.Cleanup(func() { config.SetCurrent(previous) })
}

// The task checks the plan in its own transaction, for one preview, before it
// clones anything; refused, it tells the app's notification target once the
// transaction is over, as a failed deployment with the plan's sentence.
func TestCreatePreviewIsCheckedInTheTasksTransaction(t *testing.T) {
	useDashboard(t)
	lic := beeLicense()
	lic.checkErr = hperrors.Wrap(hperrors.ErrPlanLimitReached)
	notifications := &planNotifications{notification: &entity.Notification{}}
	clone := &cloneAfterCheck{lic: lic}
	s := &service{
		appRepo:             noPreviews{},
		appCloneService:     clone,
		appService:          &previewParentApps{app: previewParentApp(t, planNotification)},
		licenseService:      lic,
		notificationService: notifications,
	}
	execData := previewTaskExecData(t, "main")

	_, err := s.CreatePreview(context.Background(), database.Tx{},
		&apppreviewservice.CreatePreviewReq{TaskExecData: execData})

	assert.ErrorIs(t, err, hperrors.ErrPlanLimitReached)
	assert.IsType(t, database.Tx{}, lic.checkDB)
	assert.Equal(t, licenseservice.Adding(base.LicenseLimitPreviews, 1), lic.checkReq)
	assert.False(t, clone.called)

	// Nothing is sent while the transaction, and the plan-limit lock, are held.
	assert.Empty(t, notifications.sent)
	if !assert.NotNil(t, execData.OnPostTxFunc) {
		t.FailNow()
	}
	execData.OnPostTxFunc()

	if !assert.Len(t, notifications.sent, 1) {
		t.FailNow()
	}
	assert.Equal(t, planNotification, notifications.eventSetting)
	if !assert.NotNil(t, notifications.eventSuccess) {
		t.FailNow()
	}
	assert.False(t, *notifications.eventSuccess)
	sent := notifications.sent[0]
	assert.False(t, sent.ActionSucceeded)
	assert.Equal(t, notificationservice.TemplateAppDeploymentNotification, sent.TemplateName)
	msg, ok := sent.TemplateData.(*notificationservice.TemplateDataAppDeployment)
	if !assert.True(t, ok) {
		t.FailNow()
	}
	assert.Equal(t, "[shop][web] Preview refused", msg.GetTitle())
	assert.False(t, msg.Succeeded)
	assert.Equal(t, "Bee runs 2 previews at a time: close one, or upgrade.", msg.Reason)
	assert.Equal(t, base.DeploymentMethodRepo, msg.Method)
	assert.Equal(t, "https://github.com/acme/web", msg.RepoURL)
	assert.Equal(t, "main", msg.RepoRef)
	// The app's previews, where one is closed. The domain is the installation's.
	assert.True(t, strings.HasSuffix(msg.DashboardLink, "/projects/proj-1/prod/apps/app-1/preview-deployments"),
		msg.DashboardLink)
}

// A preview the plan allows goes on to be made, checked before anything is
// cloned; and a failure that is not the plan's reads no license and tells no
// target here.
func TestCreatePreviewGoesOnWhenThePlanAllows(t *testing.T) {
	lic := beeLicense()
	clone := &cloneAfterCheck{lic: lic}
	s := &service{
		appRepo:             noPreviews{},
		appCloneService:     clone,
		appService:          &previewParentApps{app: previewParentApp(t, planNotification)},
		licenseService:      lic,
		notificationService: &planNotifications{},
	}
	execData := previewTaskExecData(t, "main")

	_, err := s.CreatePreview(context.Background(), database.Tx{},
		&apppreviewservice.CreatePreviewReq{TaskExecData: execData})

	assert.ErrorIs(t, err, errCloneReached)
	assert.True(t, clone.checkedFirst)
	assert.Zero(t, lic.currentCalls)
	assert.Nil(t, execData.OnPostTxFunc)
}

func TestPreviewPlanLimitOf(t *testing.T) {
	ctx := context.Background()
	refusal := hperrors.Wrap(hperrors.ErrPlanLimitReached)

	t.Run("another failure reads no license", func(t *testing.T) {
		lic := beeLicense()
		s := &service{licenseService: lic}
		assert.Nil(t, s.previewPlanLimitOf(ctx, hperrors.Wrap(hperrors.ErrAppNotFound)))
		assert.Zero(t, lic.currentCalls)
	})
	t.Run("the plan's name and limit", func(t *testing.T) {
		s := &service{licenseService: beeLicense()}
		assert.Equal(t, &previewPlanLimit{PlanName: "Bee", Max: 2}, s.previewPlanLimitOf(ctx, refusal))
	})
	t.Run("through a panic joined to it", func(t *testing.T) {
		s := &service{licenseService: beeLicense()}
		assert.NotNil(t, s.previewPlanLimitOf(ctx, errors.Join(refusal, hperrors.ErrPanic)))
	})
	t.Run("no license read, no limit to name", func(t *testing.T) {
		lic := beeLicense()
		lic.currentErr = hperrors.Wrap(hperrors.ErrInternal)
		s := &service{licenseService: lic}
		assert.Nil(t, s.previewPlanLimitOf(ctx, refusal))
	})
	t.Run("a key lifting the limit since", func(t *testing.T) {
		lic := beeLicense()
		lic.state.Plan, lic.state.Limits = base.LicensePlanHive, license.Limits{}
		s := &service{licenseService: lic}
		assert.Nil(t, s.previewPlanLimitOf(ctx, refusal))
	})
}

func TestNotifyPlanLimitRefused(t *testing.T) {
	ctx := context.Background()
	bee := &previewPlanLimit{PlanName: "Bee", Max: 2}
	newData := func(t *testing.T, app *entity.App) *createPreviewData {
		t.Helper()
		return &createPreviewData{
			CreatePreviewReq: &apppreviewservice.CreatePreviewReq{
				TaskExecData: previewTaskExecData(t, "pull/42"),
				App:          app,
			},
			Args: &entity.TaskAppPreviewArgs{RepoRef: "pull/42"},
		}
	}

	t.Run("another failure tells nobody", func(t *testing.T) {
		data := newData(t, previewParentApp(t, planNotification))
		(&service{}).notifyPlanLimitRefused(ctx, data, nil)
		assert.Nil(t, data.OnPostTxFunc)
	})
	t.Run("an app that names no target for deployments tells nobody", func(t *testing.T) {
		data := newData(t, previewParentApp(t, nil))
		(&service{}).notifyPlanLimitRefused(ctx, data, bee)
		assert.Nil(t, data.OnPostTxFunc)
	})
	t.Run("a target that resolves to nothing sends nothing", func(t *testing.T) {
		notifications := &planNotifications{}
		data := newData(t, previewParentApp(t, planNotification))
		(&service{notificationService: notifications}).notifyPlanLimitRefused(ctx, data, bee)
		if !assert.NotNil(t, data.OnPostTxFunc) {
			t.FailNow()
		}
		data.OnPostTxFunc()
		assert.NotNil(t, notifications.eventSetting)
		assert.Empty(t, notifications.sent)
	})
	t.Run("the target is told of the pull request's ref", func(t *testing.T) {
		useDashboard(t)
		notifications := &planNotifications{notification: &entity.Notification{}}
		data := newData(t, previewParentApp(t, planNotification))
		(&service{notificationService: notifications}).notifyPlanLimitRefused(ctx, data, bee)
		if !assert.NotNil(t, data.OnPostTxFunc) {
			t.FailNow()
		}
		data.OnPostTxFunc()
		if !assert.Len(t, notifications.sent, 1) {
			t.FailNow()
		}
		msg, ok := notifications.sent[0].TemplateData.(*notificationservice.TemplateDataAppDeployment)
		if !assert.True(t, ok) {
			t.FailNow()
		}
		assert.Equal(t, "pull/42", msg.RepoRef)
	})
}
