package commandpipeexecserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandservice"
)

func (s *service) calcCommandEnv(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	command *entity.CommandTemplate,
	data *execData,
) (env []string, err error) {
	resp, err := s.commandService.BuildCommand(ctx, db, &commandservice.BuildCommandReq{
		Scope:      app.GetObjectScope(),
		Command:    command,
		RefObjects: data.RefObjects,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	env = make([]string, 0, len(resp.EnvVars))
	for _, v := range resp.EnvVars {
		env = append(env, v.ToString("="))
	}

	if data.LogStore != nil && len(resp.EnvVars) > 0 {
		secrets := make(map[string]struct{}, 10) //nolint:mnd
		for _, env := range resp.EnvVars {
			for secret := range env.RefSecrets {
				plainSecret, err := secret.Value.GetPlain()
				if err != nil {
					return nil, hperrors.Wrap(err)
				}
				secrets[plainSecret] = struct{}{}
			}
		}
		data.LogStore.UpdateRedactorAddSecrets(gofn.MapKeys(secrets))
	}

	return env, nil
}
