package appsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appsettingsuc/appsettingsdto"
)

// GetAppNetworkSettings Gets app network settings
// @Summary Gets app network settings
// @Description Gets app network settings
// @Tags    app_settings
// @Produce json
// @Id      getAppNetworkSettings
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Success 200 {object} appsettingsdto.GetAppNetworkSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/network-settings [get]
func (h *Handler) GetAppNetworkSettings(ctx *gin.Context) {
	auth, projectID, appID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appsettingsdto.NewGetAppNetworkSettingsReq()
	req.ProjectID = projectID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appSettingsUC.GetAppNetworkSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateAppNetworkSettings Updates app network settings
// @Summary Updates app network settings
// @Description Updates app network settings
// @Tags    app_settings
// @Produce json
// @Id      updateAppNetworkSettings
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   body body appsettingsdto.UpdateAppNetworkSettingsReq true "request data"
// @Success 200 {object} appsettingsdto.UpdateAppNetworkSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/network-settings [put]
func (h *Handler) UpdateAppNetworkSettings(ctx *gin.Context) {
	auth, projectID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appsettingsdto.NewUpdateAppNetworkSettingsReq()
	req.ProjectID = projectID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appSettingsUC.UpdateAppNetworkSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
