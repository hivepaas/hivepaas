package appsettingsuc

import (
	"context"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) UpdateAppRoutingSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppRoutingSettingsReq,
) (*appsettingsdto.UpdateAppRoutingSettingsResp, error) {
	var data *updateAppRoutingSettingsData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data = &updateAppRoutingSettingsData{}
		err := uc.loadAppRoutingSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		err = uc.ensureDomainCerts(ctx, db, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.prepareUpdatingAppRoutingSettings(ctx, data, persistingData)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		err = uc.applyAppRoutingSettings(ctx, db, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return uc.recordAppUpdate(ctx, db, auth, data.App, base.AuditLogSourceAPIUpdate, "routing", auditdetail.New().
			WithChangedFields(parseAppSetting(data.RoutingSetting), data.NewRoutingSettings))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// A task can be picked up only once its row exists, which is once the
	// transaction has committed. A failure here is not the caller's problem: the
	// settings are saved, and the queue's own scan finds the task at its next
	// pass - what is lost is how soon.
	_ = uc.taskQueue.ScheduleTask(ctx, data.CertTasks...)

	resp := &appsettingsdto.UpdateAppRoutingSettingsResp{
		Meta: &basedto.Meta{},
	}

	if err = uc.applyAppEnvVars(ctx, uc.db, data, false); err != nil {
		// NOTE: just show user a message instead of failing the request?
		resp.Meta.Warning = "Configuration updated successfully, but failed to apply env var changes:\n" +
			err.Error()
	}

	return resp, nil
}

type updateAppRoutingSettingsData struct {
	App                *entity.App
	RoutingSetting     *entity.Setting
	NewRoutingSettings *entity.AppRoutingSettings
	RefObjects         *entity.RefObjects

	PortChanged   bool
	DomainChanged bool

	// CertTasks obtain the certificates the new domains have none for. They are
	// scheduled after the transaction commits, because that is when their rows
	// exist for the worker to claim.
	CertTasks []*entity.Task
}

func (uc *UC) loadAppRoutingSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appsettingsdto.UpdateAppRoutingSettingsReq,
	data *updateAppRoutingSettingsData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppRouting),
		),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.App = app
	data.RoutingSetting = app.GetSettingByType(base.SettingTypeAppRouting)

	if data.RoutingSetting != nil && data.RoutingSetting.UpdateVer != req.UpdateVer {
		return hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}

	newRoutingSettings := req.ToEntity()
	data.NewRoutingSettings = newRoutingSettings

	// Make sure all reference settings used in these settings exist actively
	err = uc.settingService.LoadRefObjectsByIDs(ctx, db, &data.RefObjects, app.GetObjectScope(),
		true, newRoutingSettings.GetRefObjectIDs())
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Active domains of the app need to validate
	activeDomains := newRoutingSettings.GetActiveDomainNames()

	// Verify domains are allowed in project
	err = uc.domainService.VerifyProjectDomains(ctx, db, app.ProjectID, activeDomains)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Make sure all domains used by the app are not hold by any other app
	err = uc.domainService.VerifyDomainsAvailable(ctx, db, activeDomains, []string{app.ID})
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Detect some changes
	// The app's address as a whole, not the name alone: HIVEPAAS_APP_URL carries
	// the scheme as well, so turning HTTPS on for a domain that keeps its name
	// changes what the app has to be told about itself.
	var oldPort, newPort int
	var oldURL string
	if data.RoutingSetting != nil {
		oldRoutingSettings := data.RoutingSetting.MustAsAppRoutingSettings()
		oldPort = oldRoutingSettings.Port
		oldURL = oldRoutingSettings.GetAppURL()
	}
	newPort = newRoutingSettings.Port
	data.PortChanged = oldPort != newPort
	data.DomainChanged = oldURL != newRoutingSettings.GetAppURL()

	return nil
}

// ensureDomainCerts gives each domain nobody picked a certificate for one that
// covers it, and starts obtaining what the system does not hold yet.
//
// Choosing here is what makes a domain somebody has just typed in served over
// TLS: a wildcard covering it is attached and applied with this change, and
// anything that has to be obtained arrives as a task, which attaches it and
// applies the routing again. A domain left without one is not an error - the app
// is still routed, over plain HTTP until a certificate exists.
func (uc *UC) ensureDomainCerts(
	ctx context.Context,
	db database.Tx,
	data *updateAppRoutingSettingsData,
) error {
	activeDomains := data.NewRoutingSettings.GetActiveDomains()
	var wanted []string
	for _, domain := range activeDomains {
		if domain.SSLCert.ID == "" {
			wanted = append(wanted, domain.Domain)
		}
	}
	if len(wanted) == 0 {
		return nil
	}

	app := data.App
	resp, err := uc.domainService.EnsureCertsForDomains(ctx, db, &domainservice.EnsureCertsReq{
		Scope:     app.GetObjectScope(),
		ProjectID: app.ProjectID,
		AppID:     app.ID,
		Domains:   wanted,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.CertTasks = resp.Tasks

	for _, domain := range activeDomains {
		cert := resp.Matched[domain.Domain]
		if domain.SSLCert.ID != "" || cert == nil {
			continue
		}
		domain.SSLCert = entity.ObjectID{ID: cert.ID}
		data.RefObjects.RefSettings[cert.ID] = cert
	}
	return nil
}

func (uc *UC) prepareUpdatingAppRoutingSettings(
	_ context.Context,
	data *updateAppRoutingSettingsData,
	persistingData *persistingAppData,
) {
	app := data.App
	setting := data.RoutingSetting
	timeNow := timeutil.NowUTC()

	if setting == nil {
		setting = &entity.Setting{
			ID:          gofn.Must(ulid.NewStringULID()),
			Scope:       base.ObjectScopeApp,
			ObjectID:    app.ID,
			Type:        base.SettingTypeAppRouting,
			Inheritable: true,
			CreatedAt:   timeNow,
			Version:     entity.CurrentAppRoutingSettingsVersion,
		}
		data.RoutingSetting = setting
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	setting.Status = base.SettingStatusActive
	setting.ExpireAt = time.Time{}
	setting.MustSetData(data.NewRoutingSettings)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)
}

func (uc *UC) applyAppRoutingSettings(
	ctx context.Context,
	db database.IDB,
	data *updateAppRoutingSettingsData,
) error {
	routingSettings, err := data.RoutingSetting.AsAppRoutingSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}

	_, err = uc.appRoutingService.ApplyRoutingSettings(ctx, db, &approutingservice.ApplyAppRoutingReq{
		App:             data.App,
		RoutingSettings: routingSettings,
		RefObjects:      data.RefObjects,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (uc *UC) applyAppEnvVars(
	ctx context.Context,
	db database.IDB,
	data *updateAppRoutingSettingsData,
	inTx bool,
) error {
	if !data.PortChanged && !data.DomainChanged {
		return nil
	}

	transaction := !inTx // When in Tx, must not open new transactions
	concurrency := !inTx // When in Tx, concurrency may cause runtime crash

	// When port and domain change, we need to update env vars of the apps in the env
	// as port/domain are in shared env vars.

	// Loads all apps in the env
	apps, _, err := uc.appRepo.List(ctx, db, data.App.ProjectID, nil,
		bunex.SelectWhere("app.project_env_id = ?", data.App.ProjectEnvID),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	projectEnv := data.App.ProjectEnv
	projectEnv.Project = data.App.Project
	projectEnv.Apps = apps

	affectingAppEnvData, err := uc.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db,
		projectEnv.GetObjectScope(), false, nil, transaction, concurrency)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Apply the changes of env vars to the related apps
	errMap := uc.envVarService.ApplyEnvVarsForApps(ctx, db, affectingAppEnvData, transaction, concurrency)
	if len(errMap) == 0 {
		return nil
	}

	var warning string
	for i, e := range errMap {
		warning += fmt.Sprintf("\nApp '%v': %v", affectingAppEnvData[i].App.Name, e.Error())
	}

	return hperrors.Wrap(hperrors.ErrActionFailed).WithExtraDetail("%s", warning)
}
