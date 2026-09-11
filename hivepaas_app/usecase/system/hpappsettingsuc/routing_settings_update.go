package hpappsettingsuc

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

func (uc *UC) UpdateRoutingSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.UpdateRoutingSettingsReq,
) (*hpappsettingsdto.UpdateRoutingSettingsResp, error) {
	var data *updateRoutingSettingsData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data = &updateRoutingSettingsData{}
		err := uc.loadRoutingSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.prepareUpdatingRoutingSettings(ctx, data, persistingData)

		// Inside the same transaction as the change, so a committed change always
		// has a committed deadline. There is no window in which one exists
		// without the other.
		err = uc.armProbation(ctx, db, auth,
			&probationArgs{
				AppID:       data.App.ID,
				Setting:     data.RoutingSetting,
				Snapshot:    data.Snapshot,
				Window:      data.ProbationWindow,
				SettleDelay: data.SettleDelay,
			},
			&data.probationResult,
			func(task *entity.Task) {
				persistingData.UpsertingTasks = append(persistingData.UpsertingTasks, task)
			},
		)
		if err != nil {
			return hperrors.Wrap(err)
		}

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		err = uc.applyRoutingSettings(ctx, db, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return uc.recordHivePaaSSettingsUpdate(ctx, db, auth, auditSectionRouting, auditdetail.New().
			Set("domainChanged", data.DomainChanged).
			Set("onProbation", data.probationResult.Probation != nil).
			WithChangedFields(
				settingsSnapshotData(base.SettingTypeAppRouting, data.Snapshot),
				data.NewRoutingSettings))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	uc.scheduleProbation(ctx, probationOf(data))

	if data != nil && data.DomainChanged {
		// Publish a message to reload config in other instances
		_ = uc.systemEventBus.Publish(ctx, base.SystemEventHivepaasDomainReload)
		config.SetAppDomainToNeedReload()
	}

	return &hpappsettingsdto.UpdateRoutingSettingsResp{
		Data: hpappsettingsdto.TransformPendingChange(data.Probation),
	}, nil
}

type updateRoutingSettingsData struct {
	App                *entity.App
	RoutingSetting     *entity.Setting
	NewRoutingSettings *entity.AppRoutingSettings
	RefObjects         *entity.RefObjects
	DomainChanged      bool

	// Snapshot is the settings as they were before this request touched them: the
	// state a revert restores. Taken before ApplyTo, because after it the parsed
	// settings are already the new ones.
	Snapshot        entity.SettingSnapshot
	SettleDelay     time.Duration
	ProbationWindow time.Duration

	// probationResult holds the scheduled undo of this change - see armProbation.
	probationResult
}

// probationOf survives a transaction that never got as far as building one.
func probationOf(data *updateRoutingSettingsData) *probationResult {
	if data == nil {
		return nil
	}
	return &data.probationResult
}

type persistingAppData struct {
	appservice.PersistingAppData
}

func (uc *UC) loadRoutingSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *hpappsettingsdto.UpdateRoutingSettingsReq,
	data *updateRoutingSettingsData,
) error {
	app, err := uc.hpAppService.LoadAppByKey(ctx, db, base.HivepaasAppKey,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
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

	data.Snapshot = entity.SettingSnapshotOf(data.RoutingSetting)
	// Nothing restarts: a routing change rewrites swarm service labels and waits
	// for traefik to poll. The shared floor is exactly that wait.
	data.SettleDelay = entity.SettingsProbationSettleDelay
	data.ProbationWindow = resolveProbationWindow(req.ConfirmWindow.ToDuration(), data.SettleDelay)

	routingSettings := data.RoutingSetting.MustAsAppRoutingSettings()
	var currDomain string // active domain before any change
	if domains := routingSettings.GetActiveDomainNames(); len(domains) > 0 {
		currDomain = domains[0]
	}

	if err := req.ApplyTo(routingSettings); err != nil {
		return hperrors.Wrap(err)
	}
	data.NewRoutingSettings = routingSettings

	// Checked on the result, not on the request, so it sees what will actually be
	// written - including the parts of the current settings the request left alone.
	if err := ensureStillReachable(ctx, routingSettings); err != nil {
		return hperrors.Wrap(err)
	}

	// Make sure all reference settings used in these settings exist actively
	err = uc.settingService.LoadRefObjectsByIDs(ctx, db, &data.RefObjects, app.GetObjectScope(),
		true, routingSettings.GetRefObjectIDs())
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Active domains of the app need to validate
	newActiveDomains := routingSettings.GetActiveDomainNames()

	// Verify domains are allowed in project
	err = uc.domainService.VerifyProjectDomains(ctx, db, app.ProjectID, newActiveDomains)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Make sure all domains used by the app are not hold by any other app
	err = uc.domainService.VerifyDomainsAvailable(ctx, db, newActiveDomains, []string{app.ID})
	if err != nil {
		return hperrors.Wrap(err)
	}

	if len(newActiveDomains) > 0 && newActiveDomains[0] != currDomain {
		data.DomainChanged = true
	}

	return nil
}

func (uc *UC) prepareUpdatingRoutingSettings(
	_ context.Context,
	data *updateRoutingSettingsData,
	persistingData *persistingAppData,
) {
	setting := data.RoutingSetting
	timeNow := timeutil.NowUTC()

	uc.hpAppService.SetupRoutingSettingsDefault(data.NewRoutingSettings)

	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	setting.Status = base.SettingStatusActive
	setting.ExpireAt = time.Time{}
	setting.MustSetData(data.NewRoutingSettings)
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)
}

func (uc *UC) applyRoutingSettings(
	ctx context.Context,
	db database.IDB,
	data *updateRoutingSettingsData,
) error {
	routingSettings, err := data.RoutingSetting.AsAppRoutingSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}

	resp, err := uc.appRoutingService.ApplyRoutingSettings(ctx, db, &approutingservice.ApplyAppRoutingReq{
		App:                 data.App,
		RoutingSettings:     routingSettings,
		RefObjects:          data.RefObjects,
		SkipUpdatingService: true,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	service := resp.Service
	if service.Spec.UpdateConfig == nil {
		service.Spec.UpdateConfig = &swarm.UpdateConfig{}
	}
	service.Spec.UpdateConfig.FailureAction = swarm.UpdateFailureActionRollback
	service.Spec.UpdateConfig.MaxFailureRatio = 0.5

	_, err = uc.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

func (uc *UC) persistData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingAppData,
) error {
	err := uc.appService.PersistAppData(ctx, db, &persistingData.PersistingAppData)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
