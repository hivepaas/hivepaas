package projectenvsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type GetProjectEnvEnvVarsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
}

func NewGetProjectEnvEnvVarsReq() *GetProjectEnvEnvVarsReq {
	return &GetProjectEnvEnvVarsReq{}
}

func (req *GetProjectEnvEnvVarsReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetProjectEnvEnvVarsResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *EnvVarsResp  `json:"data"`
}

type EnvVarsResp struct {
	InheritedBuildtimeEnvVars []*basedto.EnvVarResp `json:"inheritedBuildtimeEnvVars"`
	BuildtimeEnvVars          []*basedto.EnvVarResp `json:"buildtimeEnvVars"`
	InheritedRuntimeEnvVars   []*basedto.EnvVarResp `json:"inheritedRuntimeEnvVars"`
	RuntimeEnvVars            []*basedto.EnvVarResp `json:"runtimeEnvVars"`

	UpdateVer int `json:"updateVer"`
}

type EnvVarsTransformationInput struct {
	ProjectEnv *entity.ProjectEnv
	Vars       []*entity.Setting
}

func TransformEnvVars(input *EnvVarsTransformationInput) (resp *EnvVarsResp, err error) {
	resp = &EnvVarsResp{
		InheritedBuildtimeEnvVars: make([]*basedto.EnvVarResp, 0, 20), //nolint
		BuildtimeEnvVars:          make([]*basedto.EnvVarResp, 0, 20), //nolint
		InheritedRuntimeEnvVars:   make([]*basedto.EnvVarResp, 0, 20), //nolint
		RuntimeEnvVars:            make([]*basedto.EnvVarResp, 0, 20), //nolint
	}

	var envEnvVars, projectEnvVars *entity.EnvVars
	for _, envSetting := range input.Vars {
		switch envSetting.ObjectID {
		case input.ProjectEnv.ID:
			envEnvVars = envSetting.MustAsEnvVars()
		case input.ProjectEnv.ProjectID:
			projectEnvVars = envSetting.MustAsEnvVars()
		}
	}

	TransformInheritedEnvVars(projectEnvVars, resp)
	TransformOwnEnvVars(envEnvVars, resp)

	return resp, nil
}

func TransformOwnEnvVars(
	appEnvVars *entity.EnvVars,
	resp *EnvVarsResp,
) {
	if appEnvVars == nil {
		return
	}
	for _, env := range appEnvVars.Data {
		envResp := basedto.TransformEnvVar(env)
		switch {
		case env.IsBuild:
			resp.BuildtimeEnvVars = append(resp.BuildtimeEnvVars, envResp)
		default:
			resp.RuntimeEnvVars = append(resp.RuntimeEnvVars, envResp)
		}
	}
}

func TransformInheritedEnvVars(
	projectEnvVars *entity.EnvVars,
	resp *EnvVarsResp,
) {
	if projectEnvVars == nil {
		return
	}
	for _, env := range projectEnvVars.Data {
		envResp := basedto.TransformEnvVar(env)
		if env.IsBuild {
			resp.InheritedBuildtimeEnvVars = append(resp.InheritedBuildtimeEnvVars, envResp)
		} else {
			resp.InheritedRuntimeEnvVars = append(resp.InheritedRuntimeEnvVars, envResp)
		}
	}
}
