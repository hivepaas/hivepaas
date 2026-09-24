package dockerapiserviceimpl

import (
	"context"
	"slices"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
)

// nanoPerCPU is how docker counts processor time: in billionths of a CPU.
const nanoPerCPU = 1_000_000_000

// Policies leave out the apps in host mode: they talk to the node's socket, and
// an agent serving them would only collect their children less soon.
func (s *service) Policies(ctx context.Context, db database.IDB) ([]*dockerproxy.Policy, error) {
	granted, err := s.grantedAccess(ctx, db)
	if err != nil {
		return nil, err
	}
	policies := make([]*dockerproxy.Policy, 0, len(granted))
	for _, one := range granted {
		if one.access.IsHostMode() {
			continue
		}
		envNetwork := s.networkService.GetProjectNetworkName(one.app.Project, one.app.ProjectEnv.Name)
		policies = append(policies, policyOf(one.app, one.access, envNetwork))
	}
	return policies, nil
}

func (s *service) HostModeApps(ctx context.Context, db database.IDB) ([]*entity.App, error) {
	granted, err := s.grantedAccess(ctx, db)
	if err != nil {
		return nil, err
	}
	var apps []*entity.App
	for _, one := range granted {
		if one.access.IsHostMode() {
			apps = append(apps, one.app)
		}
	}
	return apps, nil
}

// appAccess is one app with Docker API access, and the access.
type appAccess struct {
	app    *entity.App
	access *entity.AppDockerAPISettings
}

// grantedAccess is every app with Docker API access, each with its project and
// env.
func (s *service) grantedAccess(ctx context.Context, db database.IDB) ([]appAccess, error) {
	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDockerAPI),
		bunex.SelectWhere("setting.scope = ?", base.ObjectScopeApp),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	appIDs := make([]string, 0, len(settings))
	for _, setting := range settings {
		appIDs = append(appIDs, setting.ObjectID)
	}
	apps, err := s.appRepo.ListByIDs(ctx, db, "", appIDs,
		bunex.SelectRelation("Project"), bunex.SelectRelation("ProjectEnv"))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	appByID := make(map[string]*entity.App, len(apps))
	for _, app := range apps {
		appByID[app.ID] = app
	}

	granted := make([]appAccess, 0, len(settings))
	for _, setting := range settings {
		// An app deleted a moment ago can still have its setting row.
		app := appByID[setting.ObjectID]
		if app == nil || app.Project == nil || app.ProjectEnv == nil {
			continue
		}
		access, err := setting.AsAppDockerAPISettings()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		granted = append(granted, appAccess{app: app, access: access})
	}
	return granted, nil
}

// policyOf is what the proxy enforces for one app.
func policyOf(app *entity.App, data *entity.AppDockerAPISettings, envNetwork string) *dockerproxy.Policy {
	policy := &dockerproxy.Policy{
		AppID:        app.ID,
		ServiceID:    app.ServiceID,
		Images:       data.Images,
		SharedDirs:   data.SharedDirs,
		Network:      dockerapiservice.NetworkName(app.ID),
		SocketVolume: dockerapiservice.SocketVolumeName(app.ID),
		Limits: dockerproxy.Limits{
			Containers: gofn.Coalesce(data.Limits.Containers, dockerapiservice.DefaultContainers),
			Memory:     gofn.Coalesce(data.Limits.Memory.Bytes(), dockerapiservice.DefaultMemory),
			NanoCPUs:   gofn.Coalesce(int64(data.Limits.CPUs*nanoPerCPU), dockerapiservice.DefaultNanoCPUs),
		},
	}
	if slices.Contains(data.Networks, entity.DockerAPINetworkEnv) {
		policy.Networks = []string{envNetwork}
	}
	for _, group := range data.Allow {
		policy.Allow = append(policy.Allow, dockerproxy.Group(group))
	}
	return policy
}
