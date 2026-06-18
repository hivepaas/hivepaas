package appdto

import (
	"context"

	"github.com/moby/moby/client"
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

var (
	SupportedShells = []string{"sh", "bash", "zsh", "fish"}
)

type OpenTerminalReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
	Shell     string `json:"-" mapstructure:"shell"`
	Width     uint   `json:"-" mapstructure:"w"`
	Height    uint   `json:"-" mapstructure:"h"`
}

func NewOpenTerminalReq() *OpenTerminalReq {
	return &OpenTerminalReq{}
}

func (req *OpenTerminalReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Shell, false, SupportedShells, "shell")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type OpenTerminalResp struct {
	Meta             *basedto.Meta                              `json:"meta"`
	ExecAttachResult *client.ExecAttachResult                   `json:"-"`
	ExecResizeFunc   func(ctx context.Context, w, h uint) error `json:"-"`
	CloseFunc        func()                                     `json:"-"`
}
