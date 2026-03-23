package apphandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

// GetAppServiceSettings Gets app service settings
// @Summary Gets app service settings
// @Description Gets app service settings
// @Tags    apps
// @Produce json
// @Id      getAppServiceSettings
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Success 200 {object} appdto.GetAppServiceSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/service-settings [get]
func (h *AppHandler) GetAppServiceSettings(ctx *gin.Context) {
	auth, projectID, appID, err := h.getAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewGetAppServiceSettingsReq()
	req.ProjectID = projectID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appUC.GetAppServiceSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateAppServiceSettings Updates app service settings
// @Summary Updates app service settings
// @Description Updates app service settings
// @Tags    apps
// @Produce json
// @Id      updateAppServiceSettings
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   body body appdto.UpdateAppServiceSettingsReq true "request data"
// @Success 200 {object} appdto.UpdateAppServiceSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/service-settings [put]
func (h *AppHandler) UpdateAppServiceSettings(ctx *gin.Context) {
	auth, projectID, appID, err := h.getAuth(ctx, base.ActionTypeWrite, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewUpdateAppServiceSettingsReq()
	req.ProjectID = projectID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appUC.UpdateAppServiceSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
