package appmanageduc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

// loadedManaged is a managed app as its endpoints need it: the app with its
// app-managed setting and its jobs, the setting parsed, and its engine.
type loadedManaged struct {
	app     *entity.App
	setting *entity.Setting
	managed *entity.AppManagedSettings
	engine  managedmodel.Engine
	jobs    []*entity.Setting
}

// loadManaged loads a managed app, and refuses one no engine looks after as not
// found: it has no managed state to read or change.
func (uc *UC) loadManaged(
	ctx context.Context,
	db database.IDB,
	projectID, appID string,
	extraOpts ...bunex.SelectQueryOption,
) (*loadedManaged, error) {
	opts := append([]bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		// LoadApp checks the project and its environment are active: it reads
		// them from these relations, and without them refuses any app.
		bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeAppManaged, base.SettingTypeSchedJob),
			// A restored database's jobs are disabled until it has recovered.
			bunex.SelectWhereIn("setting.status IN (?)", base.SettingStatusActive, base.SettingStatusDisabled),
		),
	}, extraOpts...)
	app, err := uc.appService.LoadApp(ctx, db, projectID, appID, true, true, opts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	out := &loadedManaged{app: app}
	for _, setting := range app.Settings {
		if setting.Type == base.SettingTypeAppManaged && setting.DeletedAt.IsZero() {
			out.setting = setting
		}
	}
	if out.setting == nil {
		return nil, hperrors.NewNotFound("Managed engine")
	}
	if out.managed, err = out.setting.AsAppManagedSettings(); err != nil {
		return nil, hperrors.Wrap(err)
	}
	engine, found := uc.managedService.Engine(templatemodel.ManagedEngine(out.managed.Engine))
	if !found {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateIncompatible).
			WithExtraDetail("engine %s is not one this HivePaaS has", out.managed.Engine)
	}
	out.engine = engine
	for _, setting := range app.Settings {
		if setting.Type != base.SettingTypeSchedJob {
			continue
		}
		job, err := setting.AsSchedJob()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if job.ManagedBy.ID == out.setting.ID {
			out.jobs = append(out.jobs, setting)
		}
	}
	return out, nil
}
