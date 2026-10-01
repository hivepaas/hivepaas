package apppreviewserviceimpl

import (
	"context"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

// WithheldSecrets names the secrets an app's previews go without, and the app's
// variables, runtime and build, that use them. A preview inherits only the
// secrets marked inheritable; a variable built from another comes down to it
// empty. Those of the app's env and project it does not have either way.
func (s *service) WithheldSecrets(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
) ([]*apppreviewservice.WithheldSecret, error) {
	if app.Project == nil || app.ProjectEnv == nil {
		var err error
		app, err = s.appService.LoadApp(ctx, db, app.ProjectID, app.ID, false, false,
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			bunex.SelectRelation("Project",
				bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
			),
			bunex.SelectRelation("ProjectEnv"),
		)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	var envVars []*envvarservice.EnvVar
	for _, buildPhase := range []bool{false, true} {
		resp, err := s.envVarService.BuildEnvVarsInApp(ctx, db, &envvarservice.BuildEnvVarsInAppReq{
			App:          app,
			LoadOptions:  envvarservice.EnvLoadOptions{BuildPhase: buildPhase},
			BuildOptions: envvarservice.EnvBuildOptions{BuildPhaseOnly: buildPhase, MaskSecrets: true},
		})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		envVars = append(envVars, resp.EnvVars...)
	}
	return withheldSecrets(envVars), nil
}

// withheldSecrets gathers the secrets not inheritable that the variables use,
// by name, each with the variables that use it.
func withheldSecrets(envVars []*envvarservice.EnvVar) []*apppreviewservice.WithheldSecret {
	byName := map[string]*apppreviewservice.WithheldSecret{}
	for _, v := range envVars {
		for _, setting := range v.RefSecretSettings {
			if setting.Inheritable {
				continue
			}
			secret := byName[setting.Name]
			if secret == nil {
				secret = &apppreviewservice.WithheldSecret{Name: setting.Name}
				byName[setting.Name] = secret
			}
			if !slices.Contains(secret.EnvVars, v.Key) {
				secret.EnvVars = append(secret.EnvVars, v.Key)
			}
		}
	}

	res := make([]*apppreviewservice.WithheldSecret, 0, len(byName))
	for _, secret := range byName {
		slices.Sort(secret.EnvVars)
		res = append(res, secret)
	}
	slices.SortFunc(res, func(a, b *apppreviewservice.WithheldSecret) int {
		return strings.Compare(a.Name, b.Name)
	})
	return res
}
