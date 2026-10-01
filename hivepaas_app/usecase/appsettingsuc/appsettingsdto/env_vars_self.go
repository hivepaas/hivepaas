package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice/envself"
)

type GetEnvSelfSuggestionsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	// Engine is the engine to suggest for; App Kind's when not given.
	Engine string `json:"-" mapstructure:"engine"`
}

func NewGetEnvSelfSuggestionsReq() *GetEnvSelfSuggestionsReq {
	return &GetEnvSelfSuggestionsReq{}
}

func (req *GetEnvSelfSuggestionsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Engine, false,
		gofn.MapSlice(envself.Engines, func(e *envself.Engine) string { return e.ID }), "engine")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetEnvSelfSuggestionsResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *EnvSelfSuggestionsResp `json:"data"`
}

type EnvSelfSuggestionsResp struct {
	// Engines are those suggestions are made for.
	Engines []*EnvSelfEngineResp `json:"engines"`
	// AppEngine is the engine of Engines App Kind names, or empty.
	AppEngine string `json:"appEngine"`
	// Engine is the engine suggested for: the one asked, else App Kind's; nil
	// when neither is one of Engines.
	Engine *EnvSelfEngineResp `json:"engine"`
	Vars   []*EnvLinkVarResp  `json:"vars"`
	// Command is what to run the image with, for an engine that reads no
	// variable, such as Redis.
	Command  string   `json:"command,omitempty"`
	Warnings []string `json:"warnings"`
}

type EnvSelfEngineResp struct {
	ID       string           `json:"id"`
	Title    string           `json:"title"`
	Category base.AppCategory `json:"category"`
	Image    string           `json:"image"`
	// InitOnly says the image reads the variables only when it creates its data.
	InitOnly bool `json:"initOnly"`
}

func TransformEnvSelfEngine(engine *envself.Engine) *EnvSelfEngineResp {
	if engine == nil {
		return nil
	}
	return &EnvSelfEngineResp{ID: engine.ID, Title: engine.Title, Category: engine.Category, Image: engine.Image,
		InitOnly: engine.InitOnly}
}

func TransformEnvSelfSuggestion(appEngine *envself.Engine, s *envself.Suggestion) *EnvSelfSuggestionsResp {
	resp := &EnvSelfSuggestionsResp{
		Engines:  gofn.MapSlice(envself.Engines, TransformEnvSelfEngine),
		Vars:     []*EnvLinkVarResp{},
		Warnings: []string{},
	}
	if appEngine != nil {
		resp.AppEngine = appEngine.ID
	}
	if s == nil {
		return resp
	}
	resp.Engine = TransformEnvSelfEngine(s.Engine)
	for _, v := range s.Vars {
		resp.Vars = append(resp.Vars, &EnvLinkVarResp{Key: v.Key, Value: v.Value, Description: v.Description})
	}
	resp.Command = s.Command
	resp.Warnings = append(resp.Warnings, s.Warnings...)
	return resp
}
