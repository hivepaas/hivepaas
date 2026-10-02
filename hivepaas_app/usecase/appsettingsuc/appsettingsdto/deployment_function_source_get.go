package appsettingsdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

// DeploymentFunctionSourceResp is a function's source as its deployment
// settings answer it: what DeploymentFunctionSourceReq takes back, the settings
// it refers to shown as settings.
type DeploymentFunctionSourceResp struct {
	Runtime        base.FunctionRuntime      `json:"runtime"`
	Contract       base.FunctionContract     `json:"contract"`
	Entrypoint     *FunctionEntrypointResp   `json:"entrypoint"`
	Code           *FunctionCodeResp         `json:"code"`
	SystemPackages []string                  `json:"systemPackages"`
	Timeout        timeutil.Duration         `json:"timeout" swaggertype:"string"`
	MaxConcurrency int                       `json:"maxConcurrency"`
	MaxBodySize    unit.DataSize             `json:"maxBodySize" swaggertype:"string"`
	PushToRegistry *settings.BaseSettingResp `json:"pushToRegistry"`
}

type FunctionEntrypointResp struct {
	File    string `json:"file"`
	Handler string `json:"handler"`
}

type FunctionCodeResp struct {
	Inline *FunctionInlineCodeResp `json:"inline,omitempty"`
	Repo   *FunctionRepoCodeResp   `json:"repo,omitempty"`
	Dir    string                  `json:"dir,omitempty"`
}

type FunctionInlineCodeResp struct {
	Files []*FunctionFileResp `json:"files"`
}

type FunctionFileResp struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type FunctionRepoCodeResp struct {
	RepoType    base.RepoType              `json:"repoType"`
	RepoID      string                     `json:"repoId"`
	RepoURL     string                     `json:"repoURL"`
	RepoRef     string                     `json:"repoRef"`
	CommitHash  string                     `json:"commitHash"`
	RepoOptions *DeploymentRepoOptionsResp `json:"repoOptions"`
	Credentials *settings.BaseSettingResp  `json:"credentials"`
	AutoDeploy  bool                       `json:"autoDeploy"`
}

// transformFunctionSource shows the settings a function's source refers to as
// settings: a registry, the repository's credentials.
func transformFunctionSource(resp *DeploymentFunctionSourceResp, refObjects *entity.RefObjects) {
	if resp == nil {
		return
	}
	resp.PushToRegistry = refSettingResp(resp.PushToRegistry, refObjects, base.SettingTypeRegistryAuth)
	if resp.Code != nil && resp.Code.Repo != nil {
		resp.Code.Repo.Credentials = refSettingResp(resp.Code.Repo.Credentials, refObjects,
			base.SettingTypeAccessToken)
	}
}

// refSettingResp is a setting referred to, as a response shows it: nothing when
// none is set, and missing when it is gone.
func refSettingResp(
	ref *settings.BaseSettingResp,
	refObjects *entity.RefObjects,
	typ base.SettingType,
) *settings.BaseSettingResp {
	if ref == nil || ref.ID == "" {
		return nil
	}
	itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[ref.ID])
	if itemResp == nil {
		itemResp = settings.NewMissingSetting(ref.ID, typ)
	}
	return itemResp
}
