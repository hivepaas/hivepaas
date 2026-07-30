package envvarservice

import (
	"context"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type ParseErrorType string

const (
	ParseErrorSecretMissing ParseErrorType = "secret-missing"
	ParseErrorSecretFailure ParseErrorType = "secret-failure"
	ParseErrorVarMissing    ParseErrorType = "var-missing"
)

type ParseError struct {
	Type ParseErrorType
	Name string
}

func (e *ParseError) Error() string {
	switch e.Type {
	case ParseErrorSecretMissing:
		return fmt.Sprintf("secret '%v' is missing", e.Name)
	case ParseErrorSecretFailure:
		return fmt.Sprintf("secret '%v' is failed to parse", e.Name)
	case ParseErrorVarMissing:
		return fmt.Sprintf("variable '%v' is missing", e.Name)
	default:
		return fmt.Sprintf("unknown error '%v' at '%v'", e.Type, e.Name)
	}
}

func (e *ParseError) ErrorWithApp(app string) string {
	return "App '" + app + "': " + e.Error()
}

type EnvVar struct {
	*entity.EnvVar
	RefSecrets map[*entity.Secret]struct{}
	Errors     []*ParseError
}

func (env *EnvVar) ToString(sep string) string {
	return env.Key + sep + env.Value
}

func (env *EnvVar) AddRefSecret(secret *entity.Secret) {
	if env.RefSecrets == nil {
		env.RefSecrets = make(map[*entity.Secret]struct{})
	}
	env.RefSecrets[secret] = struct{}{}
}

type EnvVarsData struct {
	EnvVars []*EnvVar
	Secrets []*entity.Setting
}

type EnvLoadOptions struct {
	BuildPhase bool
}

type EnvLoadFunc func(context.Context, database.IDB, *base.ObjectScope, EnvLoadOptions) (
	[]*EnvVar, []*entity.Setting, error)

func NewStaticEnvLoadFunc(vars []*EnvVar, secrets []*entity.Setting) EnvLoadFunc {
	return func(context.Context, database.IDB, *base.ObjectScope, EnvLoadOptions) ([]*EnvVar, []*entity.Setting, error) {
		return vars, secrets, nil
	}
}

type AppEnvVarData struct {
	App     *entity.App
	EnvVars []*EnvVar
	Secrets []*entity.Setting
}

func (e *AppEnvVarData) Errors() (res []string) {
	for _, env := range e.EnvVars {
		for _, err := range env.Errors {
			res = append(res, err.ErrorWithApp(e.App.Name))
		}
	}
	return res
}

type EnvBuildOptions struct {
	SkipLoadingVars    bool
	SkipLoadingSecrets bool
	MaskSecrets        bool
	BuildPhaseOnly     bool
	SharedVarsOnly     bool
	Sort               bool
}

type BuildEnvVarsInAppReq struct {
	App                   *entity.App
	TargetVars            []string
	OverridingVars        []*EnvVar
	DataLoadFunc          EnvLoadFunc // if nil, use default loader
	InheritedDataLoadFunc EnvLoadFunc // if nil, use default loader
	LoadOptions           EnvLoadOptions
	BuildOptions          EnvBuildOptions
}

type BuildEnvVarsInAppResp struct {
	EnvVars          []*EnvVar
	Secrets          []*entity.Setting
	InheritedEnvVars []*EnvVar
	InheritedSecrets []*entity.Setting
}

type BuildSystemEnvVarsInAppReq struct {
	App  *entity.App
	Sort bool
}

type BuildEnvVarsInProjectReq struct {
	Project        *entity.Project
	TargetVars     []string
	OverridingVars []*EnvVar
	DataLoadFunc   EnvLoadFunc // if nil, use default loader
	LoadOptions    EnvLoadOptions
	BuildOptions   EnvBuildOptions
}

type BuildEnvVarsInProjectResp struct {
	EnvVars []*EnvVar
	Secrets []*entity.Setting
}

type BuildSystemEnvVarsInProjectReq struct {
	Project *entity.Project
	Sort    bool
}

type BuildEnvVarsInProjectEnvReq struct {
	ProjectEnv            *entity.ProjectEnv
	TargetVars            []string
	OverridingVars        []*EnvVar
	DataLoadFunc          EnvLoadFunc // if nil, use default loader
	InheritedDataLoadFunc EnvLoadFunc // if nil, use default loader
	LoadOptions           EnvLoadOptions
	BuildOptions          EnvBuildOptions
}

type BuildEnvVarsInProjectEnvResp struct {
	EnvVars          []*EnvVar
	Secrets          []*entity.Setting
	InheritedEnvVars []*EnvVar
	InheritedSecrets []*entity.Setting
}

type BuildSystemEnvVarsInProjectEnvReq struct {
	ProjectEnv *entity.ProjectEnv
	Sort       bool
}
