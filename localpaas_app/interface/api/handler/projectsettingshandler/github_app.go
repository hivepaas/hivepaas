package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/githubappuc/githubappdto"
)

// ListGithubApp Lists github-app settings
// @Summary Lists github-app settings
// @Description Lists github-app settings
// @Tags    project_settings
// @Produce json
// @Id      listProjectGithubApp
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} githubappdto.ListGithubAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/github-apps [get]
func (h *Handler) ListGithubApp(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeGithubApp, base.SettingScopeProject)
}

// GetGithubApp Gets github-app setting details
// @Summary Gets github-app setting details
// @Description Gets github-app setting details
// @Tags    project_settings
// @Produce json
// @Id      getProjectGithubApp
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} githubappdto.GetGithubAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/github-apps/{itemID} [get]
func (h *Handler) GetGithubApp(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeGithubApp, base.SettingScopeProject)
}

// CreateGithubApp Creates a new github-app setting
// @Summary Creates a new github-app setting
// @Description Creates a new github-app setting
// @Tags    project_settings
// @Produce json
// @Id      createProjectGithubApp
// @Param   projectID path string true "project ID"
// @Param   body body githubappdto.CreateGithubAppReq true "request data"
// @Success 201 {object} githubappdto.CreateGithubAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/github-apps [post]
func (h *Handler) CreateGithubApp(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeGithubApp, base.SettingScopeProject)
}

// UpdateGithubApp Updates github-app
// @Summary Updates github-app
// @Description Updates github-app
// @Tags    project_settings
// @Produce json
// @Id      updateProjectGithubApp
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body githubappdto.UpdateGithubAppReq true "request data"
// @Success 200 {object} githubappdto.UpdateGithubAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/github-apps/{itemID} [put]
func (h *Handler) UpdateGithubApp(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeGithubApp, base.SettingScopeProject)
}

// UpdateGithubAppStatus Updates github-app status
// @Summary Updates github-app status
// @Description Updates github-app status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectGithubAppStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body githubappdto.UpdateGithubAppStatusReq true "request data"
// @Success 200 {object} githubappdto.UpdateGithubAppStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/github-apps/{itemID}/status [put]
func (h *Handler) UpdateGithubAppStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeGithubApp, base.SettingScopeProject)
}

// DeleteGithubApp Deletes github-app setting
// @Summary Deletes github-app setting
// @Description Deletes github-app setting
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectGithubApp
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} githubappdto.DeleteGithubAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/github-apps/{itemID} [delete]
func (h *Handler) DeleteGithubApp(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeGithubApp, base.SettingScopeProject)
}

// BeginProjectGithubAppManifestFlow Begins a github-app manifest flow
// @Summary Begins a github-app manifest flow
// @Description Begins a github-app manifest flow
// @Tags    project_settings
// @Produce json
// @Id      beginProjectGithubAppManifestFlow
// @Param   projectID path string true "project ID"
// @Param   body body githubappdto.BeginGithubAppManifestFlowReq true "request data"
// @Success 200 {object} githubappdto.BeginGithubAppManifestFlowResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/github-apps/manifest-flow/begin [post]
func (h *Handler) BeginProjectGithubAppManifestFlow(ctx *gin.Context) {
	h.GithubAppManifestFlowBegin(ctx, base.SettingScopeProject)
}

// BeginProjectGithubAppManifestFlowCreation Begins a github-app manifest flow creation
// @Summary Begins a github-app manifest flow creation
// @Description Begins a github-app manifest flow creation
// @Tags    project_settings
// @Produce json
// @Id      beginProjectGithubAppManifestFlowCreation
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 "html page to redirect to github app creation page"
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/github-apps/{itemID}/manifest-flow/begin [get]
func (h *Handler) BeginProjectGithubAppManifestFlowCreation(ctx *gin.Context) {
	h.GithubAppManifestFlowBeginCreation(ctx, base.SettingScopeProject)
}

// HandleProjectGithubAppManifestFlowProgress Handles progress of github-app manifest flow
// @Summary Handles progress of github-app manifest flow
// @Description Handles progress of github-app manifest flow
// @Tags    project_settings
// @Produce json
// @Id      handleProjectGithubAppManifestFlowProgress
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 "html page to redirect to github app creation page"
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/github-apps/itemID}/manifest-flow/progress [get]
func (h *Handler) HandleProjectGithubAppManifestFlowProgress(ctx *gin.Context) {
	h.GithubAppManifestFlowProgress(ctx, base.SettingScopeProject)
}

// BeginReprovisionProjectGithubApp Begins reprovisioning a github-app through manifest flow
// @Summary Begins reprovisioning a github-app through manifest flow
// @Description Begins reprovisioning a github-app through manifest flow
// @Tags    project_settings
// @Produce json
// @Id      beginReprovisionProjectGithubApp
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body githubappdto.BeginReprovisionGithubAppReq true "request data"
// @Success 200 {object} githubappdto.BeginReprovisionGithubAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/github-apps/{itemID}/begin-reprovision [post]
func (h *Handler) BeginReprovisionProjectGithubApp(ctx *gin.Context) {
	h.GithubAppBeginReprovision(ctx, base.SettingScopeProject)
}
