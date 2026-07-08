package schedjobserviceimpl

import (
	"context"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) BuildCommandEnv(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	schedJob *entity.SchedJob,
) (res []*envvarservice.EnvVar, usedSecrets []*entity.Secret, err error) {
	envVars := schedJob.Command.EnvVars

	for _, argGroup := range schedJob.Command.ArgGroups {
		if env := s.buildEnvForArgs(argGroup); env != nil {
			envVars = append(envVars, env)
		}
	}

	// Quick check to see if we need to load secrets
	loadSecrets := false
	for _, env := range envVars {
		if !env.IsLiteral && s.envVarService.HasSecretRef(env.Value) {
			loadSecrets = true
			break
		}
	}

	res, usedSecrets, err = s.envVarService.ProcessEnvRefs(ctx, db, app, envVars,
		false, loadSecrets, false)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}
	return res, usedSecrets, nil
}

func (s *service) buildEnvForArgs(
	argGroup *entity.CommandTemplateArgGroup,
) *entity.EnvVar {
	if !argGroup.Enabled || len(argGroup.Args) == 0 {
		return nil
	}

	buf := &strings.Builder{}
	buf.Grow(100) //nolint:mnd
	separator := gofn.Coalesce(argGroup.Separator, " ")

	for _, arg := range argGroup.Args {
		if !arg.Use {
			continue
		}
		if buf.Len() > 0 {
			buf.WriteString(" ")
		}
		if arg.Value == "" {
			buf.WriteString(arg.Name)
		} else {
			buf.WriteString(arg.Name + separator + executil.ArgQuote(arg.Value))
		}
	}
	if buf.Len() == 0 {
		return nil
	}
	return &entity.EnvVar{Key: argGroup.ExportEnv, Value: buf.String()}
}
