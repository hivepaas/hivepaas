package basedto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type EnvVarReq struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	IsLiteral bool   `json:"isLiteral"`
}

func (req *EnvVarReq) ToEntity(kind base.EnvVarKind) *entity.EnvVar {
	return &entity.EnvVar{
		Key:       req.Key,
		Value:     req.Value,
		IsLiteral: req.IsLiteral,
		IsBuild:   kind == base.EnvVarKindBuild,
		IsShared:  kind == base.EnvVarKindShared,
	}
}

type EnvVarResp struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	IsLiteral  bool   `json:"isLiteral,omitempty"`
	IsSystem   bool   `json:"isSystem,omitempty"`
	IsReadOnly bool   `json:"isReadOnly,omitempty"`
}

func TransformEnvVar(env *entity.EnvVar) *EnvVarResp {
	return &EnvVarResp{
		Key:       env.Key,
		Value:     env.Value,
		IsLiteral: env.IsLiteral,
		IsSystem:  env.IsSystem,
	}
}
