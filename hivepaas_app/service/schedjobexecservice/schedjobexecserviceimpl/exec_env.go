package schedjobexecserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func (s *service) calcCommandEnv(
	ctx context.Context,
	db database.Tx,
	data *execData,
) (env []string, err error) {
	schedJob := data.SchedJobSetting.MustAsSchedJob()
	envVars, refSecrets, err := s.schedJobService.BuildCommandEnv(ctx, db, data.App, schedJob)
	if err != nil {
		return nil, apperrors.New(err)
	}

	env = make([]string, 0, len(envVars))
	for _, v := range envVars {
		env = append(env, v.ToString("="))
	}

	if len(refSecrets) > 0 && data.LogStore != nil {
		secrets := make([]string, 0, len(refSecrets))
		for _, secret := range refSecrets {
			plainSecret, err := secret.Value.GetPlain()
			if err != nil {
				return nil, apperrors.New(err)
			}
			secrets = append(secrets, plainSecret)
		}
		data.LogStore.UpdateRedactorAddSecrets(secrets)
	}
	return env, nil
}
