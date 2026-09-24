package hpappsettingsdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetSecuritySettingsReq struct {
}

func NewGetSecuritySettingsReq() *GetSecuritySettingsReq {
	return &GetSecuritySettingsReq{}
}

func (req *GetSecuritySettingsReq) Validate() hperrors.ValidationErrors {
	return nil
}

type GetSecuritySettingsResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *SecuritySettingsResp `json:"data"`
}

type SecuritySettingsResp struct {
	ReturnSecretsViaAPI     bool     `json:"returnSecretsViaApi"`
	AlwaysReturnSecretTypes []string `json:"alwaysReturnSecretTypes"`
	AllowPrivilegedApps     bool     `json:"allowPrivilegedApps"`
	// PrivilegedApps are the apps given the node's own Docker socket. Turning the
	// switch off takes it from none of them: this is where an administrator
	// finds them.
	PrivilegedApps []*PrivilegedAppResp `json:"privilegedApps"`
}

// PrivilegedAppResp is an app given the node's own Docker socket, with where to
// find it: the dashboard's routes name an env by its key.
type PrivilegedAppResp struct {
	AppID          string `json:"appId"`
	AppName        string `json:"appName"`
	ProjectID      string `json:"projectId"`
	ProjectName    string `json:"projectName"`
	ProjectEnvKey  string `json:"projectEnvKey"`
	ProjectEnvName string `json:"projectEnvName"`
}

type SecuritySettingsTransformInput struct {
	Config *config.Config
	// PrivilegedApps are loaded with their Project and ProjectEnv.
	PrivilegedApps []*entity.App
}

func TransformSecuritySettings(input *SecuritySettingsTransformInput) (resp *SecuritySettingsResp, err error) {
	// Normalised to an empty list rather than passed through as nil: the field is
	// a set the client renders, and "no exemptions" arriving as null makes every
	// reader handle a case that carries no extra meaning.
	exemptions := input.Config.Security.AlwaysReturnSecretTypes
	if exemptions == nil {
		exemptions = []string{}
	}
	resp = &SecuritySettingsResp{
		ReturnSecretsViaAPI:     input.Config.Security.ReturnSecretsViaAPI,
		AlwaysReturnSecretTypes: exemptions,
		AllowPrivilegedApps:     input.Config.Security.AllowPrivilegedApps,
		PrivilegedApps:          make([]*PrivilegedAppResp, 0, len(input.PrivilegedApps)),
	}
	for _, app := range input.PrivilegedApps {
		resp.PrivilegedApps = append(resp.PrivilegedApps, &PrivilegedAppResp{
			AppID:          app.ID,
			AppName:        app.Name,
			ProjectID:      app.ProjectID,
			ProjectName:    app.Project.Name,
			ProjectEnvKey:  app.ProjectEnv.Key,
			ProjectEnvName: app.ProjectEnv.Name,
		})
	}
	return resp, nil
}
