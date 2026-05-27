package schedjobserviceimpl

import (
	"context"
	"strings"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/shellutil"
	"github.com/localpaas/localpaas/localpaas_app/service/envvarservice"
)

func (s *service) BuildCommandEnv(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	schedJob *entity.SchedJob,
) (res []*envvarservice.EnvVar, err error) {
	envVars := schedJob.Command.EnvVars

	for _, argGroup := range schedJob.Command.ArgGroups {
		if env := s.buildEnvForArgs(argGroup); env != nil {
			envVars = append(envVars, env)
		}
	}

	// Quick check to see if we need to replace all references in the ENV values
	needReplaceRefs := false
	for _, env := range envVars {
		if strings.Contains(env.Value, "${secrets.") {
			needReplaceRefs = true
			break
		}
	}

	if needReplaceRefs {
		res, err = s.envVarService.ProcessEnvRefs(ctx, db, app, envVars, false, true, false)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		return res, nil
	}

	res = make([]*envvarservice.EnvVar, 0, len(envVars))
	for _, env := range envVars {
		res = append(res, &envvarservice.EnvVar{EnvVar: env})
	}
	return res, nil
}

func (s *service) buildEnvForArgs(
	argGroup *entity.SchedJobCommandArgGroup,
) *entity.EnvVar {
	if len(argGroup.Args) == 0 {
		return nil
	}
	buf := &strings.Builder{}
	buf.Grow(100) //nolint:mnd
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
			buf.WriteString(arg.Name + argGroup.Separator + shellutil.ArgQuote(arg.Value))
		}
	}
	if buf.Len() == 0 {
		return nil
	}
	return &entity.EnvVar{Key: argGroup.ExportEnv, Value: buf.String()}
}
