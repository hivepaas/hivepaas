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

func (s *service) Policies(ctx context.Context, db database.IDB) ([]*dockerproxy.Policy, error) {
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

	policies := make([]*dockerproxy.Policy, 0, len(settings))
	for _, setting := range settings {
		// An app deleted a moment ago can still have its setting row.
		app := appByID[setting.ObjectID]
		if app == nil || app.Project == nil || app.ProjectEnv == nil {
			continue
		}
		data, err := setting.AsAppDockerAPISettings()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		envNetwork := s.networkService.GetProjectNetworkName(app.Project, app.ProjectEnv.Name)
		policies = append(policies, policyOf(app, data, envNetwork))
	}
	return policies, nil
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
			Memory:     gofn.Coalesce(data.Limits.Memory, dockerapiservice.DefaultMemory),
			NanoCPUs:   gofn.Coalesce(data.Limits.NanoCPUs, dockerapiservice.DefaultNanoCPUs),
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
