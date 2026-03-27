package projectdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

type GetProjectEnvVarsReq struct {
	ProjectID string `json:"-"`
}

func NewGetProjectEnvVarsReq() *GetProjectEnvVarsReq {
	return &GetProjectEnvVarsReq{}
}

func (req *GetProjectEnvVarsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetProjectEnvVarsResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *EnvVarsResp  `json:"data"`
}

type EnvVarsResp struct {
	BuildtimeEnvVars []*basedto.EnvVarResp `json:"buildtimeEnvVars"`
	RuntimeEnvVars   []*basedto.EnvVarResp `json:"runtimeEnvVars"`
	UpdateVer        int                   `json:"updateVer"`
}

func TransformEnvVars(setting *entity.Setting) (resp *EnvVarsResp, err error) {
	if setting == nil {
		return
	}
	resp = &EnvVarsResp{
		BuildtimeEnvVars: make([]*basedto.EnvVarResp, 0, 20), //nolint
		RuntimeEnvVars:   make([]*basedto.EnvVarResp, 0, 20), //nolint
		UpdateVer:        setting.UpdateVer,
	}

	envVars, err := setting.AsEnvVars()
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if envVars != nil {
		for _, v := range envVars.Data {
			res := basedto.TransformEnvVar(v)
			if v.IsBuildEnv {
				resp.BuildtimeEnvVars = append(resp.BuildtimeEnvVars, res)
			} else {
				resp.RuntimeEnvVars = append(resp.RuntimeEnvVars, res)
			}
		}
	}
	return resp, nil
}
