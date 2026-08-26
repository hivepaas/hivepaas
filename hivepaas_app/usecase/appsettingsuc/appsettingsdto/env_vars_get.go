package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

type GetAppEnvVarsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetAppEnvVarsReq() *GetAppEnvVarsReq {
	return &GetAppEnvVarsReq{}
}

func (req *GetAppEnvVarsReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppEnvVarsResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *EnvVarsResp  `json:"data"`
}

type EnvVarsResp struct {
	InheritedBuildtimeEnvVars []*basedto.EnvVarResp `json:"inheritedBuildtimeEnvVars"`
	BuildtimeEnvVars          []*basedto.EnvVarResp `json:"buildtimeEnvVars"`
	InheritedRuntimeEnvVars   []*basedto.EnvVarResp `json:"inheritedRuntimeEnvVars"`
	RuntimeEnvVars            []*basedto.EnvVarResp `json:"runtimeEnvVars"`
	SharedEnvVars             []*basedto.EnvVarResp `json:"sharedEnvVars"`
	UpdateVer                 int                   `json:"updateVer"`
}

type EnvVarsTransformationInput struct {
	App               *entity.App
	Vars              []*entity.Setting
	SystemVars        []*envvarservice.EnvVar
	ParentSystemVars  []*envvarservice.EnvVar
	EnvSystemVars     []*envvarservice.EnvVar
	ProjectSystemVars []*envvarservice.EnvVar
}

func TransformEnvVars(input *EnvVarsTransformationInput) (resp *EnvVarsResp, err error) {
	resp = &EnvVarsResp{
		InheritedBuildtimeEnvVars: make([]*basedto.EnvVarResp, 0, 20), //nolint
		BuildtimeEnvVars:          make([]*basedto.EnvVarResp, 0, 20), //nolint
		InheritedRuntimeEnvVars:   make([]*basedto.EnvVarResp, 0, 20), //nolint
		RuntimeEnvVars:            make([]*basedto.EnvVarResp, 0, 20), //nolint
		SharedEnvVars:             make([]*basedto.EnvVarResp, 0, 10), //nolint
	}

	var appVars, parentAppVars, projectEnvVars, projectVars *entity.EnvVars
	for _, envSetting := range input.Vars {
		switch envSetting.ObjectID {
		case input.App.ID:
			appVars = envSetting.MustAsEnvVars()
			resp.UpdateVer = envSetting.UpdateVer
		case input.App.ProjectEnvID:
			projectEnvVars = envSetting.MustAsEnvVars()
		case input.App.ProjectID:
			projectVars = envSetting.MustAsEnvVars()
		case input.App.ParentID:
			parentAppVars = envSetting.MustAsEnvVars()
		}
	}

	TransformInheritedEnvVars(projectVars, projectEnvVars, parentAppVars, input, resp)
	TransformOwnEnvVars(appVars, input, resp)

	return resp, nil
}

func TransformOwnEnvVars(
	appVars *entity.EnvVars,
	input *EnvVarsTransformationInput,
	resp *EnvVarsResp,
) {
	for _, env := range input.SystemVars {
		envResp := basedto.TransformEnvVar(env.EnvVar)
		switch {
		case env.IsBuild:
			resp.BuildtimeEnvVars = append(resp.BuildtimeEnvVars, envResp)
		case env.IsShared:
			resp.SharedEnvVars = append(resp.SharedEnvVars, envResp)
		default:
			resp.RuntimeEnvVars = append(resp.RuntimeEnvVars, envResp)
		}
	}
	if appVars != nil {
		for _, env := range appVars.Data {
			envResp := basedto.TransformEnvVar(env)
			switch {
			case env.IsBuild:
				resp.BuildtimeEnvVars = append(resp.BuildtimeEnvVars, envResp)
			case env.IsShared:
				resp.SharedEnvVars = append(resp.SharedEnvVars, envResp)
			default:
				resp.RuntimeEnvVars = append(resp.RuntimeEnvVars, envResp)
			}
		}
	}
}

//nolint:gocognit
func TransformInheritedEnvVars(
	projectVars, projectEnvVars, parentAppVars *entity.EnvVars,
	input *EnvVarsTransformationInput,
	resp *EnvVarsResp,
) {
	for _, env := range input.ProjectSystemVars {
		envResp := basedto.TransformEnvVar(env.EnvVar)
		if env.IsBuild {
			resp.InheritedBuildtimeEnvVars = append(resp.InheritedBuildtimeEnvVars, envResp)
		} else {
			resp.InheritedRuntimeEnvVars = append(resp.InheritedRuntimeEnvVars, envResp)
		}
	}
	if projectVars != nil {
		for _, env := range projectVars.Data {
			envResp := basedto.TransformEnvVar(env)
			if env.IsBuild {
				resp.InheritedBuildtimeEnvVars = append(resp.InheritedBuildtimeEnvVars, envResp)
			} else {
				resp.InheritedRuntimeEnvVars = append(resp.InheritedRuntimeEnvVars, envResp)
			}
		}
	}

	for _, env := range input.EnvSystemVars {
		envResp := basedto.TransformEnvVar(env.EnvVar)
		if env.IsBuild {
			resp.InheritedBuildtimeEnvVars = append(resp.InheritedBuildtimeEnvVars, envResp)
		} else {
			resp.InheritedRuntimeEnvVars = append(resp.InheritedRuntimeEnvVars, envResp)
		}
	}
	if projectEnvVars != nil {
		for _, env := range projectEnvVars.Data {
			envResp := basedto.TransformEnvVar(env)
			if env.IsBuild {
				resp.InheritedBuildtimeEnvVars = append(resp.InheritedBuildtimeEnvVars, envResp)
			} else {
				resp.InheritedRuntimeEnvVars = append(resp.InheritedRuntimeEnvVars, envResp)
			}
		}
	}

	for _, env := range input.ParentSystemVars {
		envResp := basedto.TransformEnvVar(env.EnvVar)
		if env.IsBuild {
			resp.InheritedBuildtimeEnvVars = append(resp.InheritedBuildtimeEnvVars, envResp)
		} else {
			resp.InheritedRuntimeEnvVars = append(resp.InheritedRuntimeEnvVars, envResp)
		}
	}
	if parentAppVars != nil {
		for _, env := range parentAppVars.Data {
			envResp := basedto.TransformEnvVar(env)
			if env.IsBuild {
				resp.InheritedBuildtimeEnvVars = append(resp.InheritedBuildtimeEnvVars, envResp)
			} else {
				resp.InheritedRuntimeEnvVars = append(resp.InheritedRuntimeEnvVars, envResp)
			}
		}
	}

	resp.InheritedBuildtimeEnvVars = removeDuplicatedEnvVars(resp.InheritedBuildtimeEnvVars)
	resp.InheritedRuntimeEnvVars = removeDuplicatedEnvVars(resp.InheritedRuntimeEnvVars)
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
