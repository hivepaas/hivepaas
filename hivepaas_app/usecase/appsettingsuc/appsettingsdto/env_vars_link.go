package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice/envlink"
)

type ListEnvLinkTargetsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewListEnvLinkTargetsReq() *ListEnvLinkTargetsReq {
	return &ListEnvLinkTargetsReq{}
}

func (req *ListEnvLinkTargetsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListEnvLinkTargetsResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data []*EnvLinkTargetResp `json:"data"`
}

type EnvLinkTargetResp struct {
	ID       string           `json:"id"`
	Key      string           `json:"key"`
	Name     string           `json:"name"`
	Category base.AppCategory `json:"category"`
	Engine   string           `json:"engine"`
}

type GetEnvLinkSuggestionsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	TargetAppID  string `json:"-" mapstructure:"targetAppId"`
}

func NewGetEnvLinkSuggestionsReq() *GetEnvLinkSuggestionsReq {
	return &GetEnvLinkSuggestionsReq{}
}

func (req *GetEnvLinkSuggestionsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateID(&req.TargetAppID, true, "targetAppId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetEnvLinkSuggestionsResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *EnvLinkSuggestionsResp `json:"data"`
}

type EnvLinkSuggestionsResp struct {
	Target *EnvLinkTargetResp  `json:"target"`
	Groups []*EnvLinkGroupResp `json:"groups"`
}

type EnvLinkGroupResp struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Recommended bool              `json:"recommended"`
	Warnings    []string          `json:"warnings"`
	Vars        []*EnvLinkVarResp `json:"vars"`
}

type EnvLinkVarResp struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

func TransformEnvLinkTarget(target *envlink.Target) *EnvLinkTargetResp {
	return &EnvLinkTargetResp{ID: target.ID, Key: target.Key, Name: target.Name,
		Category: target.Category, Engine: target.Engine}
}

func TransformEnvLinkGroups(groups []*envlink.Group) []*EnvLinkGroupResp {
	resp := make([]*EnvLinkGroupResp, 0, len(groups))
	for _, g := range groups {
		vars := make([]*EnvLinkVarResp, 0, len(g.Vars))
		for _, v := range g.Vars {
			vars = append(vars, &EnvLinkVarResp{Key: v.Key, Value: v.Value, Description: v.Description})
		}
		resp = append(resp, &EnvLinkGroupResp{ID: g.ID, Title: g.Title, Description: g.Description,
			Recommended: g.Recommended, Warnings: append([]string{}, g.Warnings...), Vars: vars})
	}
	return resp
}
