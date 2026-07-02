package basedto

import "github.com/hivepaas/hivepaas/hivepaas_app/entity"

type EnvVarReq struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	IsLiteral bool   `json:"isLiteral"`
}

func (req *EnvVarReq) ToEntity(isBuildEnv bool) *entity.EnvVar {
	return &entity.EnvVar{
		Key:        req.Key,
		Value:      req.Value,
		IsLiteral:  req.IsLiteral,
		IsBuildEnv: isBuildEnv,
	}
}

type EnvVarResp struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	IsLiteral bool   `json:"isLiteral,omitempty"`
}

func TransformEnvVar(env *entity.EnvVar) *EnvVarResp {
	return &EnvVarResp{
		Key:       env.Key,
		Value:     env.Value,
		IsLiteral: env.IsLiteral,
	}
}
