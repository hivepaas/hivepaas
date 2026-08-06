package commandpipeexecserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func (s *service) calcCommandEnv(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	command *entity.CommandTemplate,
	data *execData,
) (env []string, err error) {
	envVars, err := s.commandService.BuildCommandEnvVars(ctx, db, app, command)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	env = make([]string, 0, len(envVars))
	for _, v := range envVars {
		env = append(env, v.ToString("="))
	}

	if data.LogStore != nil && len(envVars) > 0 {
		secrets := make(map[string]struct{}, 10) //nolint:mnd
		for _, env := range envVars {
			for secret := range env.RefSecrets {
				plainSecret, err := secret.Value.GetPlain()
				if err != nil {
					return nil, apperrors.Wrap(err)
				}
				secrets[plainSecret] = struct{}{}
			}
		}
		data.LogStore.UpdateRedactorAddSecrets(gofn.MapKeys(secrets))
	}

	return env, nil
}
