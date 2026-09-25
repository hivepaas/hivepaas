package appsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// ListEnvLinkTargets Lists the apps an app may link its env vars to
// @Summary Lists the apps an app may link its env vars to
// @Description Lists the apps an app may link its env vars to
// @Tags    app_settings
// @Produce json
// @Id      listAppEnvLinkTargets
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Success 200 {object} appsettingsdto.ListEnvLinkTargetsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/env-vars/link-targets [get]
func (h *Handler) ListEnvLinkTargets(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appsettingsdto.NewListEnvLinkTargetsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appSettingsUC.ListEnvLinkTargets(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetEnvLinkSuggestions Suggests the env vars that link an app to another
// @Summary Suggests the env vars that link an app to another
// @Description Suggests the env vars that link an app to another
// @Tags    app_settings
// @Produce json
// @Id      getAppEnvLinkSuggestions
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   targetAppId query string true "the app to link to"
// @Success 200 {object} appsettingsdto.GetEnvLinkSuggestionsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/env-vars/link-suggestions [get]
func (h *Handler) GetEnvLinkSuggestions(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appsettingsdto.NewGetEnvLinkSuggestionsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appSettingsUC.GetEnvLinkSuggestions(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
