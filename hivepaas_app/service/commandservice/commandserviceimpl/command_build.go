package commandserviceimpl

import (
	"context"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

//nolint:gocognit
func (s *service) BuildCommand(
	ctx context.Context,
	db database.IDB,
	req *commandservice.BuildCommandReq,
) (resp *commandservice.BuildCommandResp, err error) {
	resp = &commandservice.BuildCommandResp{}
	commandTpl := req.Command

	// Loads ref objects
	refObjectIDs := commandTpl.GetRefObjectIDs()
	err = s.settingService.LoadRefObjectsByIDs(ctx, db, &req.RefObjects, req.Scope, false, refObjectIDs)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Get command/script value from the command template
	resp.CommandString = gofn.Coalesce(commandTpl.Command, commandTpl.Script.Value)
	if resp.CommandString == "" && commandTpl.Script.ID != "" {
		scriptSetting := req.RefObjects.RefSettings[commandTpl.Script.ID]
		if scriptSetting == nil {
			return nil, apperrors.NewNotFound("Command script")
		}
		resp.CommandString = scriptSetting.MustAsScript().Data
	}

	// Build env vars
	cmdVars := commandTpl.EnvVars
	for _, argGroup := range commandTpl.ArgGroups {
		if env := s.buildEnvVarForArgs(argGroup); env != nil {
			cmdVars = append(cmdVars, env)
		}
	}

	targetVarKeys := make([]string, 0, len(cmdVars))
	targetVars := make([]*envvarservice.EnvVar, 0, len(cmdVars))
	for _, env := range cmdVars {
		targetVarKeys = append(targetVarKeys, env.Key)
		targetVars = append(targetVars, &envvarservice.EnvVar{EnvVar: env})
	}

	var finalCmdKey string
	if req.BuildFinalCommand {
		finalCmdKey = "CMD_STR_" + gofn.RandString(16) //nolint:mnd
		targetVarKeys = append(targetVarKeys, finalCmdKey)
		targetVars = append(targetVars, &envvarservice.EnvVar{EnvVar: &entity.EnvVar{
			Key: finalCmdKey, Value: resp.CommandString,
		}})
	}

	hasRef := false
	hasSecretRef := false
	for _, env := range targetVars {
		if env.IsLiteral {
			continue
		}
		if !hasRef && s.envVarService.HasRef(env.Value) {
			hasRef = true
		}
		if !hasSecretRef && s.envVarService.HasSecretRef(env.Value) {
			hasSecretRef = true
		}
		if hasRef && hasSecretRef {
			break
		}
	}
	if hasRef || hasSecretRef {
		envResp, err := s.envVarService.BuildEnvVarsInApp(ctx, db, &envvarservice.BuildEnvVarsInAppReq{
			App:            req.Scope.App,
			TargetVars:     targetVarKeys,
			OverridingVars: targetVars,
			DataLoadFunc:   s.loadAppEnvsAndSecrets,
			LoadOptions: envvarservice.EnvLoadOptions{
				SkipLoadingSecrets: !hasSecretRef,
			},
		})
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		resp.EnvVars = envResp.EnvVars
	} else {
		resp.EnvVars = targetVars
	}

	if req.BuildFinalCommand {
		tempVars := resp.EnvVars
		resp.EnvVars = make([]*envvarservice.EnvVar, 0, len(tempVars))
		for _, env := range tempVars {
			if env.Key == finalCmdKey {
				resp.CommandString = env.Value
				continue
			}
			resp.EnvVars = append(resp.EnvVars, env)
		}
	}

	return resp, nil
}

//nolint:gocognit
func (s *service) loadAppEnvsAndSecrets(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	options envvarservice.EnvLoadOptions,
) (envVars []*envvarservice.EnvVar, secrets []*entity.Setting, err error) {
	app := scope.App
	if !scope.IsAppScope() || app.ServiceID == "" {
		envVars, secrets, err = s.envVarService.DefaultEnvLoad(ctx, db, scope, options)
		if err != nil {
			return nil, nil, apperrors.Wrap(err)
		}
		return envVars, secrets, nil
	}

	// Gets envs from swarm service for consistency and performance
	inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	envs := inspect.Service.Spec.TaskTemplate.ContainerSpec.Env
	envVars = make([]*envvarservice.EnvVar, 0, len(envs))
	for _, env := range envs {
		k, v, _ := strings.Cut(env, "=")
		if k == "" {
			continue
		}
		envVars = append(envVars, &envvarservice.EnvVar{
			EnvVar: &entity.EnvVar{Key: k, Value: v},
		})
	}

	// Loads secrets if instructed
	if !options.SkipLoadingSecrets {
		settings, _, err := s.settingRepo.List(ctx, db, scope, nil,
			bunex.SelectWhere("setting.type = ?", base.SettingTypeSecret),
			bunex.SelectWhere("setting.size <= ?", base.SecretRefInEnvMaxSize),
			bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		)
		if err != nil {
			return nil, nil, apperrors.Wrap(err)
		}
		secretMap := make(map[string]*entity.Setting, len(settings))
		for _, setting := range settings { // project secrets
			if setting.ObjectID == app.ProjectID {
				secretMap[setting.Name] = setting
			}
		}
		for _, setting := range settings { // project env secrets
			if setting.ObjectID == app.ProjectEnvID {
				secretMap[setting.Name] = setting
			}
		}
		if app.ParentID != "" {
			for _, setting := range settings { // parent app secrets
				if setting.ObjectID == app.ParentID {
					secretMap[setting.Name] = setting
				}
			}
		}
		for _, setting := range settings { // current app secrets
			if setting.ObjectID == app.ID {
				secretMap[setting.Name] = setting
			}
		}
		secrets = gofn.MapValues(secretMap)
	}

	return envVars, secrets, nil
}

func (s *service) buildEnvVarForArgs(
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
