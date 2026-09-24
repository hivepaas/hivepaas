package appsettingsuc

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// UpdateAppDockerAPISettings saves what an app may do through the Docker API.
// The setting is written in the transaction; the agents, the app's service and
// its environment follow after the commit, the way the env vars screen applies
// its changes, and what fails there is a warning on a saved setting. The app's
// next deployment brings its service to the setting anyway.
func (uc *UC) UpdateAppDockerAPISettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppDockerAPISettingsReq,
) (*appsettingsdto.UpdateAppDockerAPISettingsResp, error) {
	data := &updateAppDockerAPIData{}
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		if err := uc.loadAppDockerAPIForUpdate(ctx, db, req, data); err != nil {
			return err
		}
		if err := uc.prepareAppDockerAPI(ctx, db, auth, req, data); err != nil {
			return err
		}
		if data.Setting == nil {
			return nil
		}
		persisting := &persistingAppData{}
		persisting.UpsertingSettings = append(persisting.UpsertingSettings, data.Setting)
		if err := uc.persistData(ctx, db, persisting); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.recordAppUpdate(ctx, db, auth, data.App, base.AuditLogSourceAPIUpdate, "docker-api",
			auditdetail.New().
				Set("enabled", data.Next != nil).
				Set("hostMode", data.Next.IsHostMode()).
				Set("widened", dockerapiservice.Widens(data.Prev, data.Next)))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp := &appsettingsdto.UpdateAppDockerAPISettingsResp{Meta: &basedto.Meta{}}
	if err = uc.applyAppDockerAPI(ctx, data); err != nil {
		resp.Meta.Warning = "Configuration updated successfully, but failed to apply changes:\n" + err.Error()
	}
	return resp, nil
}

type updateAppDockerAPIData struct {
	App     *entity.App
	Service *swarm.Service
	// Row is the app's setting as it was, nil when it had none. Setting is the
	// one written, nil when nothing is.
	Row     *entity.Setting
	Setting *entity.Setting
	// Prev and Next are the access the app had and has, nil for none.
	Prev *entity.AppDockerAPISettings
	Next *entity.AppDockerAPISettings
}

func (uc *UC) loadAppDockerAPIForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appsettingsdto.UpdateAppDockerAPISettingsReq,
	data *updateAppDockerAPIData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.App = app

	if data.Row, err = uc.appDockerAPISetting(ctx, db, app.ID); err != nil {
		return err
	}
	if data.Row != nil && data.Row.UpdateVer != req.UpdateVer {
		return hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}
	if data.Row != nil && data.Row.Status == base.SettingStatusActive {
		if data.Prev, err = data.Row.AsAppDockerAPISettings(); err != nil {
			return hperrors.Wrap(err)
		}
	}

	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.Service, err = requireAppService(service, app.ID)
	return err
}

// prepareAppDockerAPI decides what saving the screen writes. Turning access
// off with none to turn off writes nothing.
func (uc *UC) prepareAppDockerAPI(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppDockerAPISettingsReq,
	data *updateAppDockerAPIData,
) error {
	if req.Enabled {
		data.Next = req.ToEntity()
		if problem := dockerAPIProblem(data.Next, data.Service); problem != "" {
			return hperrors.Wrap(hperrors.ErrValueInvalid).WithExtraDetail("%s", problem)
		}
	} else if data.Row == nil {
		return nil
	}
	if err := uc.checkDockerAPIGrant(ctx, db, auth, data.Prev, data.Next); err != nil {
		return err
	}
	data.Setting = nextDockerAPISetting(data.App, data.Row, data.Next, timeutil.NowUTC())
	return nil
}

// dockerAPIProblem says what is wrong with the access asked for, empty when
// nothing is. A shared directory has to be on one of the volumes the app's
// service mounts: that volume is what a child is given. Host mode uses no
// policy, so none is held against the service.
func dockerAPIProblem(access *entity.AppDockerAPISettings, service *swarm.Service) string {
	if problem := specmodel.DockerAPIProblem(access); problem != "" || access.IsHostMode() {
		return problem
	}
	var targets []string
	if container := service.Spec.TaskTemplate.ContainerSpec; container != nil {
		for i := range container.Mounts {
			m := &container.Mounts[i]
			if m.Type == mount.TypeVolume && !dockerapiservice.IsSocketMount(m) {
				targets = append(targets, m.Target)
			}
		}
	}
	return specmodel.SharedDirsProblem(access.SharedDirs, targets)
}

// checkDockerAPIGrant refuses the node's own socket to a caller who is not an
// administrator, or while the privileged-apps switch is off, and access given,
// or widened, through the proxy to a caller without Write on the Cluster
// module: the gate a template giving it passes. Narrowing access, leaving host
// mode or turning it off takes only the app's Write, which the route asked for.
func (uc *UC) checkDockerAPIGrant(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	prev, next *entity.AppDockerAPISettings,
) error {
	if dockerapiservice.EntersHostMode(prev, next) {
		switch hostModeBlockedBy(config.Current(), auth) {
		case hostModeBlockedBySwitch:
			return hperrors.Wrap(hperrors.ErrUnauthorized).
				WithExtraDetail("giving an app the node's Docker socket needs the privileged-apps switch, " +
					"which is off; an administrator turns it on in the security settings").
				WithMsgLog("host mode requires the privileged-apps switch")
		case hostModeBlockedByAdmin:
			return hperrors.Wrap(hperrors.ErrUnauthorized).
				WithExtraDetail("giving an app the node's Docker socket needs an administrator").
				WithMsgLog("host mode requires an administrator")
		}
		return nil
	}
	if !dockerapiservice.Widens(prev, next) {
		return nil
	}
	hasPerm, err := uc.permissionManager.CheckAccess(ctx, db, auth, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleCluster,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if !hasPerm {
		return hperrors.Wrap(hperrors.ErrUnauthorized).
			WithExtraDetail("giving an app the Docker API, or letting it do more, " +
				"needs Write permission on the Cluster module").
			WithMsgLog("widening an app's Docker API requires Write on the Cluster module")
	}
	return nil
}

// nextDockerAPISetting is the row saving the screen writes: the app's own, or a
// new one, active with next's access, or disabled with the access it had kept
// when next is nil.
func nextDockerAPISetting(
	app *entity.App, row *entity.Setting, next *entity.AppDockerAPISettings, now time.Time,
) *entity.Setting {
	setting := row
	if setting == nil {
		setting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.ObjectScopeApp,
			ObjectID:  app.ID,
			Type:      base.SettingTypeAppDockerAPI,
			CreatedAt: now,
			Version:   entity.CurrentAppDockerAPIVersion,
		}
	}
	setting.UpdateVer++
	setting.UpdatedAt = now
	setting.Status = base.SettingStatusDisabled
	if next != nil {
		setting.Status = base.SettingStatusActive
		setting.MustSetData(next)
	}
	return setting
}

// applyAppDockerAPI brings the agents, the app's service and its environment to
// the access just saved.
//   - With access, the agents are synced first: the service's new task looks for
//     its socket as it starts, and many apps give up when it is not there. An app
//     entering host mode is no longer served.
//   - When access appears or goes, or its mode changes, the service gains or
//     loses its socket and network - or the node's socket - and
//     HIVEPAAS_DOCKER_HOST appears, goes, or names the other socket.
//   - When it goes, the app's children and their leftovers are removed at once.
//     The app's network stays, unused, until the app is deleted: removing it now
//     would race the task still leaving it.
func (uc *UC) applyAppDockerAPI(ctx context.Context, data *updateAppDockerAPIData) error {
	if data.Setting == nil {
		return nil
	}
	var errs []error
	if data.Next != nil {
		errs = append(errs, uc.dockerAPIService.SyncAgents(ctx))
	}
	if dockerAPIShape(data.Prev) != dockerAPIShape(data.Next) {
		errs = append(errs, uc.dockerManager.ServiceUpdateFunc(ctx, data.Service.ID, data.Service,
			func(_ int, service *swarm.Service) (bool, error) {
				return true, uc.dockerAPIService.ApplyToService(ctx, uc.db, data.App.ID, &service.Spec)
			}, defaultServiceRetryMax, 0))
		errs = append(errs, uc.applyDockerHostVar(ctx, data.App))
	}
	if data.Prev != nil && data.Next == nil {
		errs = append(errs, uc.dockerAPIService.RemoveAppObjects(ctx, data.App.ID))
	}
	return errors.Join(errs...)
}

// dockerAPIShape is what access gives a service: nothing, the proxy's socket and
// network, or the node's socket.
func dockerAPIShape(access *entity.AppDockerAPISettings) string {
	switch {
	case access == nil:
		return ""
	case access.IsHostMode():
		return entity.DockerAPIModeHost
	}
	return entity.DockerAPIModeProxy
}

// applyDockerHostVar builds an app's environment again, with or without
// HIVEPAAS_DOCKER_HOST, and applies it: a service whose environment does not
// change, because none of its variables names it, is left as it is.
func (uc *UC) applyDockerHostVar(ctx context.Context, app *entity.App) error {
	envData, err := uc.envVarService.BuildEnvVarsForAllAppsInScope(ctx, uc.db, app.GetObjectScope(),
		false, nil, true, true)
	if err != nil {
		return hperrors.Wrap(err)
	}
	var errs []error
	for _, err := range uc.envVarService.ApplyEnvVarsForApps(ctx, uc.db, envData, true, true) {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
