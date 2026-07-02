package settinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/githubappuc/githubappdto"
)

// ListGithubApp Lists github-app settings
// @Summary Lists github-app settings
// @Description Lists github-app settings
// @Tags    settings
// @Produce json
// @Id      listSettingGithubApp
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} githubappdto.ListGithubAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/github-apps [get]
func (h *Handler) ListGithubApp(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeGithubApp, base.ObjectScopeGlobal)
}

// GetGithubApp Gets github-app setting details
// @Summary Gets github-app setting details
// @Description Gets github-app setting details
// @Tags    settings
// @Produce json
// @Id      getSettingGithubApp
// @Param   itemID path string true "setting ID"
// @Success 200 {object} githubappdto.GetGithubAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/github-apps/{itemID} [get]
func (h *Handler) GetGithubApp(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeGithubApp, base.ObjectScopeGlobal)
}

// CreateGithubApp Creates a new github-app setting
// @Summary Creates a new github-app setting
// @Description Creates a new github-app setting
// @Tags    settings
// @Produce json
// @Id      createSettingGithubApp
// @Param   body body githubappdto.CreateGithubAppReq true "request data"
// @Success 201 {object} githubappdto.CreateGithubAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/github-apps [post]
func (h *Handler) CreateGithubApp(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeGithubApp, base.ObjectScopeGlobal)
}

// UpdateGithubApp Updates github-app
// @Summary Updates github-app
// @Description Updates github-app
// @Tags    settings
// @Produce json
// @Id      updateSettingGithubApp
// @Param   itemID path string true "setting ID"
// @Param   body body githubappdto.UpdateGithubAppReq true "request data"
// @Success 200 {object} githubappdto.UpdateGithubAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/github-apps/{itemID} [put]
func (h *Handler) UpdateGithubApp(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeGithubApp, base.ObjectScopeGlobal)
}

// UpdateGithubAppStatus Updates github-app status
// @Summary Updates github-app status
// @Description Updates github-app status
// @Tags    settings
// @Produce json
// @Id      updateSettingGithubAppStatus
// @Param   itemID path string true "setting ID"
// @Param   body body githubappdto.UpdateGithubAppStatusReq true "request data"
// @Success 200 {object} githubappdto.UpdateGithubAppStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/github-apps/{itemID}/status [put]
func (h *Handler) UpdateGithubAppStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeGithubApp, base.ObjectScopeGlobal)
}

// DeleteGithubApp Deletes github-app setting
// @Summary Deletes github-app setting
// @Description Deletes github-app setting
// @Tags    settings
// @Produce json
// @Id      deleteSettingGithubApp
// @Param   itemID path string true "setting ID"
// @Success 200 {object} githubappdto.DeleteGithubAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/github-apps/{itemID} [delete]
func (h *Handler) DeleteGithubApp(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeGithubApp, base.ObjectScopeGlobal)
}

// TestGithubAppConn Test github app connection
// @Summary Test github app connection
// @Description Test github app connection
// @Tags    settings
// @Produce json
// @Id      testGithubAppConn
// @Param   body body githubappdto.TestGithubAppConnReq true "request data"
// @Success 200 {object} githubappdto.TestGithubAppConnResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/github-apps/test-conn [post]
func (h *Handler) TestGithubAppConn(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := githubappdto.NewTestGithubAppConnReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.GithubAppUC.TestGithubAppConn(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ListAppInstallation List github app installation
// @Summary List github app installation
// @Description List github app installation
// @Tags    settings
// @Produce json
// @Id      listAppInstallation
// @Param   body body githubappdto.ListAppInstallationReq true "request data"
// @Success 200 {object} githubappdto.ListAppInstallationResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/github-apps/installations/list [post]
func (h *Handler) ListAppInstallation(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := githubappdto.NewListAppInstallationReq()
	if err = h.ParseRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.GithubAppUC.ListAppInstallation(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// BeginGithubAppManifestFlow Begins a github-app manifest flow
// @Summary Begins a github-app manifest flow
// @Description Begins a github-app manifest flow
// @Tags    settings
// @Produce json
// @Id      beginGithubAppManifestFlow
// @Param   body body githubappdto.BeginGithubAppManifestFlowReq true "request data"
// @Success 200 {object} githubappdto.BeginGithubAppManifestFlowResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/github-apps/manifest-flow/begin [post]
func (h *Handler) BeginGithubAppManifestFlow(ctx *gin.Context) {
	h.GithubAppManifestFlowBegin(ctx, base.ObjectScopeGlobal)
}

// BeginGithubAppManifestFlowCreation Begins a github-app manifest flow of creation
// @Summary Begins a github-app manifest flow of creation
// @Description Begins a github-app manifest flow of creation
// @Tags    settings
// @Produce json
// @Id      beginGithubAppManifestFlowCreation
// @Param   itemID path string true "setting ID"
// @Success 200 "html page to redirect to github app creation page"
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/github-apps/{itemID}/manifest-flow/begin [get]
func (h *Handler) BeginGithubAppManifestFlowCreation(ctx *gin.Context) {
	h.GithubAppManifestFlowBeginCreation(ctx, base.ObjectScopeGlobal)
}

// HandleGithubAppManifestFlowProgress Handles progress of github-app manifest flow
// @Summary Handles progress of github-app manifest flow
// @Description Handles progress of github-app manifest flow
// @Tags    settings
// @Produce json
// @Id      handleGithubAppManifestFlowProgress
// @Param   itemID path string true "setting ID"
// @Success 200 "html page to redirect to github app creation page"
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/github-apps/{itemID}/manifest-flow/progress [get]
func (h *Handler) HandleGithubAppManifestFlowProgress(ctx *gin.Context) {
	h.GithubAppManifestFlowProgress(ctx, base.ObjectScopeGlobal)
}

// BeginReprovisionGithubApp Begins reprovisioning a github-app through manifest flow
// @Summary Begins reprovisioning a github-app through manifest flow
// @Description Begins reprovisioning a github-app through manifest flow
// @Tags    settings
// @Produce json
// @Id      beginReprovisionGithubApp
// @Param   itemID path string true "setting ID"
// @Param   body body githubappdto.BeginReprovisionGithubAppReq true "request data"
// @Success 200 {object} githubappdto.BeginReprovisionGithubAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/github-apps/{itemID}/begin-reprovision [post]
func (h *Handler) BeginReprovisionGithubApp(ctx *gin.Context) {
	h.GithubAppBeginReprovision(ctx, base.ObjectScopeGlobal)
}
