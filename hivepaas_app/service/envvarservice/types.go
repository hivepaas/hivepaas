package envvarservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type EnvVar struct {
	*entity.EnvVar
	RefSecrets map[*entity.Secret]struct{}
	Errors     []string
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

type ComputeEnvVarsInAppReq struct {
	App              *entity.App
	TargetVars       []string
	OverridingVars   []*EnvVar
	InheritedVars    []*EnvVar         // if nil, data will be loaded from DB when needed
	InheritedSecrets []*entity.Setting // if nil, data will be loaded from DB when needed

	SkipLoadingVars    bool
	SkipLoadingSecrets bool
	MaskSecrets        bool
	BuildPhaseOnly     bool
	SharedVarsOnly     bool
	Sort               bool
}

type ComputeEnvVarsInAppResp struct {
	EnvVars          []*EnvVar
	Secrets          []*entity.Setting
	InheritedEnvVars []*EnvVar
	InheritedSecrets []*entity.Setting
}

type ComputeSystemEnvVarsInAppReq struct {
	App  *entity.App
	Sort bool
}

type ComputeEnvVarsInProjectReq struct {
	Project            *entity.Project
	TargetVars         []string
	OverridingVars     []*EnvVar
	SkipLoadingVars    bool
	SkipLoadingSecrets bool
	MaskSecrets        bool
	BuildPhaseOnly     bool
	SharedVarsOnly     bool
	Sort               bool
}

type ComputeEnvVarsInProjectResp struct {
	EnvVars []*EnvVar
	Secrets []*entity.Setting
}

type ComputeSystemEnvVarsInProjectReq struct {
	Project *entity.Project
	Sort    bool
}

type ComputeEnvVarsInProjectEnvReq struct {
	ProjectEnv       *entity.ProjectEnv
	TargetVars       []string
	OverridingVars   []*EnvVar
	InheritedVars    []*EnvVar         // if nil, data will be loaded from DB when needed
	InheritedSecrets []*entity.Setting // if nil, data will be loaded from DB when needed

	SkipLoadingVars    bool
	SkipLoadingSecrets bool
	MaskSecrets        bool
	BuildPhaseOnly     bool
	SharedVarsOnly     bool
	Sort               bool
}

type ComputeEnvVarsInProjectEnvResp struct {
	EnvVars          []*EnvVar
	Secrets          []*entity.Setting
	InheritedEnvVars []*EnvVar
	InheritedSecrets []*entity.Setting
}

type ComputeSystemEnvVarsInProjectEnvReq struct {
	ProjectEnv *entity.ProjectEnv
	Sort       bool
}
