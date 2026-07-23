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

type ComputeAppEnvVarsReq struct {
	App                    *entity.App
	TargetVars             []string
	OverridingVars         []*EnvVar
	InheritedProjectVars   []*EnvVar // if nil, data will be loaded from DB when needed
	InheritedParentAppVars []*EnvVar // if nil, data will be loaded from DB when needed
	SkipLoadingVars        bool
	SkipLoadingSecrets     bool
	MaskSecrets            bool
	BuildPhaseOnly         bool
	SharedVarsOnly         bool
	Sort                   bool
}

type ComputeAppSystemEnvVarsReq struct {
	App  *entity.App
	Sort bool
}

type ComputeProjectEnvVarsReq struct {
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

type ComputeProjectSystemEnvVarsReq struct {
	Project *entity.Project
	Sort    bool
}
