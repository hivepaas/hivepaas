// Package tasksslobtain gets the certificate an app's domain was promised.
//
// It is a task rather than part of saving a domain because a certificate
// authority takes its own time - seconds over an HTTP challenge, minutes over a
// DNS one while a record propagates - and because what it refuses, it refuses
// for a reason that will not have changed by the next deployment. A failure is
// written onto the certificate setting so the next attempt waits instead of
// spending another of the authority's weekly allowances.
package tasksslobtain

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/domainhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// retryAfter is how long a refusal is remembered, which is what stops a
// redeploy loop from asking an authority the same question every few minutes.
const retryAfter = 6 * time.Hour

type Executor struct {
	appRepo             repository.AppRepo
	appService          appservice.Service
	settingRepo         repository.SettingRepo
	settingMountService settingmountservice.Service
	settingService      settingservice.Service
	sslService          sslservice.Service
	appRoutingService   approutingservice.Service
	taskQueue           queue.TaskQueue
	logger              logging.Logger
}

func NewExecutor(
	taskQueue queue.TaskQueue,
	appRepo repository.AppRepo,
	appService appservice.Service,
	settingRepo repository.SettingRepo,
	settingMountService settingmountservice.Service,
	settingService settingservice.Service,
	sslService sslservice.Service,
	appRoutingService approutingservice.Service,
	logger logging.Logger,
) *Executor {
	e := &Executor{
		appRepo:             appRepo,
		appService:          appService,
		settingRepo:         settingRepo,
		settingMountService: settingMountService,
		settingService:      settingService,
		sslService:          sslService,
		appRoutingService:   appRoutingService,
		taskQueue:           taskQueue,
		logger:              logger,
	}
	taskQueue.RegisterExecutor(base.TaskTypeSSLObtain, e.execute)
	return e
}

func (e *Executor) execute(
	ctx context.Context,
	db database.Tx,
	task *queue.TaskExecData,
) error {
	args, err := task.Task.ArgsAsSSLObtain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if args == nil || args.SettingID == "" {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("ssl obtain task has no certificate")
	}

	setting, err := e.settingRepo.GetByID(ctx, db, nil, base.SettingTypeSSLCert, args.SettingID, true)
	if err != nil {
		return hperrors.Wrap(err)
	}
	cert, err := setting.AsSSLCert()
	if err != nil {
		return hperrors.Wrap(err)
	}
	output := &entity.TaskSSLObtainOutput{Domain: cert.Domain}
	defer func() { task.Task.MustSetOutput(output) }()

	refObjects := entity.NewRefObjects()
	if err = e.settingService.LoadRefObjects(ctx, db, &refObjects, settingScopeOf(setting), true, setting); err != nil {
		return hperrors.Wrap(err)
	}

	if _, err = e.sslService.ObtainCert(ctx, setting, refObjects, true); err != nil {
		output.Error = err.Error()
		if saveErr := e.rememberFailure(ctx, db, setting, err); saveErr != nil {
			return hperrors.Wrap(saveErr)
		}
		return hperrors.Wrap(err)
	}
	output.Obtained = true

	if err = e.saveObtained(ctx, db, setting); err != nil {
		return hperrors.Wrap(err)
	}
	// An app mounting this certificate gets it once the task's transaction has
	// committed.
	refresh, err := e.settingMountService.RecordRefresh(ctx, db, setting)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if refresh != nil {
		task.OnPostTx(func() { _ = e.taskQueue.ScheduleTask(context.WithoutCancel(ctx), refresh) })
	}

	// The certificate exists now; the app is still being served without it until
	// traefik is told, which is what this does. A failure here is worth a retry -
	// the certificate is already obtained, so the retry costs nothing with the
	// authority.
	if args.AppID != "" {
		if err = e.applyToApp(ctx, db, args.AppID, setting); err != nil {
			output.Error = err.Error()
			return hperrors.Wrap(err)
		}
		output.Applied = true
	}
	return nil
}

// rememberFailure writes onto the certificate why it has none yet, and how long
// the next attempt should wait.
func (e *Executor) rememberFailure(
	ctx context.Context,
	db database.Tx,
	setting *entity.Setting,
	cause error,
) error {
	cert := setting.MustAsSSLCert()
	cert.LastError = cause.Error()
	cert.RetryAfter = timeutil.NowUTC().Add(retryAfter)
	if err := setting.SetData(cert); err != nil {
		return hperrors.Wrap(err)
	}
	setting.UpdatedAt = timeutil.NowUTC()
	if err := e.settingRepo.Update(ctx, db, setting,
		bunex.UpdateColumns("data", "updated_at")); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (e *Executor) saveObtained(ctx context.Context, db database.Tx, setting *entity.Setting) error {
	cert := setting.MustAsSSLCert()
	cert.LastError = ""
	cert.RetryAfter = time.Time{}
	if err := setting.SetData(cert); err != nil {
		return hperrors.Wrap(err)
	}
	setting.UpdatedAt = timeutil.NowUTC()
	if err := e.settingRepo.Update(ctx, db, setting,
		bunex.UpdateColumns("data", "updated_at")); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// applyToApp gives the app's domains the certificate and routes it again, which
// is when traefik learns of it.
//
// Only domains that have none of their own are touched: a domain somebody chose
// a certificate for keeps it. A domain this certificate does not cover is left
// alone as well - the app may carry several, and this task obtained one.
func (e *Executor) applyToApp(
	ctx context.Context,
	db database.Tx,
	appID string,
	certSetting *entity.Setting,
) error {
	app, err := e.appRepo.GetByID(ctx, db, "", appID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	routingSetting, err := e.settingRepo.GetSingle(ctx, db, app.GetObjectScope(),
		base.SettingTypeAppRouting, true)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if routingSetting == nil {
		return nil
	}
	routing, err := routingSetting.AsAppRoutingSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}

	cert := certSetting.MustAsSSLCert()
	attached := false
	for _, domain := range routing.Domains {
		if domain == nil || !domain.Enabled || domain.SSLCert.ID != "" {
			continue
		}
		if !domainhelper.IsDomainCoveredByCert(domain.Domain, cert.Domain) {
			continue
		}
		domain.SSLCert = entity.ObjectID{ID: certSetting.ID}
		attached = true
	}
	if attached {
		routingSetting.UpdateVer++
		routingSetting.UpdatedAt = timeutil.NowUTC()
		routingSetting.MustSetData(routing)
		if err = e.appService.PersistAppData(ctx, db, &appservice.PersistingAppData{
			UpsertingSettings: []*entity.Setting{routingSetting},
		}); err != nil {
			return hperrors.Wrap(err)
		}
	}

	// The certificate is handed over rather than looked up again: the apply
	// resolves references from the app's scope, and a wildcard living at the
	// installation's is one it would have to inherit to find.
	refObjects := entity.NewRefObjects()
	refObjects.RefSettings[certSetting.ID] = certSetting
	_, err = e.appRoutingService.ApplyRoutingSettings(ctx, db, &approutingservice.ApplyAppRoutingReq{
		App:             app,
		RoutingSettings: routing,
		RefObjects:      refObjects,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func settingScopeOf(setting *entity.Setting) *entity.ObjectScope {
	if setting.Scope == base.ObjectScopeProject && setting.ObjectID != "" {
		return entity.NewObjectScopeProject(setting.ObjectID)
	}
	return entity.NewObjectScopeGlobal()
}
