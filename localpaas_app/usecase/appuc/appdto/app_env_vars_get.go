package appdto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

type GetAppEnvVarsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
}

func NewGetAppEnvVarsReq() *GetAppEnvVarsReq {
	return &GetAppEnvVarsReq{}
}

func (req *GetAppEnvVarsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppEnvVarsResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *EnvVarsResp  `json:"data"`
}

type EnvVarsResp struct {
	InheritedBuildtimeEnvVars []*basedto.EnvVarResp `json:"inheritedBuildtimeEnvVars"`
	InheritedRuntimeEnvVars   []*basedto.EnvVarResp `json:"inheritedRuntimeEnvVars"`
	BuildtimeEnvVars          []*basedto.EnvVarResp `json:"buildtimeEnvVars"`
	RuntimeEnvVars            []*basedto.EnvVarResp `json:"runtimeEnvVars"`
	UpdateVer                 int                   `json:"updateVer"`
}

func TransformEnvVars(app *entity.App, envVars []*entity.Setting) (resp *EnvVarsResp, err error) {
	if len(envVars) == 0 {
		return nil, nil
	}

	resp = &EnvVarsResp{
		InheritedBuildtimeEnvVars: make([]*basedto.EnvVarResp, 0, 20), //nolint
		InheritedRuntimeEnvVars:   make([]*basedto.EnvVarResp, 0, 20), //nolint
		BuildtimeEnvVars:          make([]*basedto.EnvVarResp, 0, 20), //nolint
		RuntimeEnvVars:            make([]*basedto.EnvVarResp, 0, 20), //nolint
	}

	var appEnvVars, parentAppEnvVars, projectEnvVars *entity.EnvVars
	for _, env := range envVars {
		switch env.ObjectID {
		case app.ID:
			appEnvVars = env.MustAsEnvVars()
			resp.UpdateVer = env.UpdateVer
		case app.ProjectID:
			projectEnvVars = env.MustAsEnvVars()
		case app.ParentID:
			parentAppEnvVars = env.MustAsEnvVars()
		}
	}

	if projectEnvVars != nil {
		for _, v := range projectEnvVars.Data {
			res := basedto.TransformEnvVar(v)
			if v.IsBuildEnv {
				resp.InheritedBuildtimeEnvVars = append(resp.InheritedBuildtimeEnvVars, res)
			} else {
				resp.InheritedRuntimeEnvVars = append(resp.InheritedRuntimeEnvVars, res)
			}
		}
	}
	if parentAppEnvVars != nil {
		for _, v := range parentAppEnvVars.Data {
			res := basedto.TransformEnvVar(v)
			if v.IsBuildEnv {
				resp.InheritedBuildtimeEnvVars = append(resp.InheritedBuildtimeEnvVars, res)
			} else {
				resp.InheritedRuntimeEnvVars = append(resp.InheritedRuntimeEnvVars, res)
			}
		}
	}
	if appEnvVars != nil {
		for _, v := range appEnvVars.Data {
			res := basedto.TransformEnvVar(v)
			if v.IsBuildEnv {
				resp.BuildtimeEnvVars = append(resp.BuildtimeEnvVars, res)
			} else {
				resp.RuntimeEnvVars = append(resp.RuntimeEnvVars, res)
			}
		}
	}

	if projectEnvVars != nil && parentAppEnvVars != nil &&
		len(projectEnvVars.Data) > 0 && len(parentAppEnvVars.Data) > 0 {
		resp.InheritedBuildtimeEnvVars = removeDuplicatedEnvVars(resp.InheritedBuildtimeEnvVars)
		resp.InheritedRuntimeEnvVars = removeDuplicatedEnvVars(resp.InheritedRuntimeEnvVars)
	}

	return resp, nil
}

func removeDuplicatedEnvVars(envVars []*basedto.EnvVarResp) (resp []*basedto.EnvVarResp) {
	resp = make([]*basedto.EnvVarResp, 0, len(envVars))
	mapSeen := make(map[string]struct{}, len(envVars))

	gofn.ForEachReverse(envVars, func(_ int, e *basedto.EnvVarResp) {
		if _, exists := mapSeen[e.Key]; !exists {
			resp = append(resp, e)
			mapSeen[e.Key] = struct{}{}
		}
	})

	return gofn.Reverse(resp)
}
