package appsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// GetAppAutoscale Gets a function's autoscale
// @Summary Gets a function's autoscale
// @Description Gets a function's autoscale: on or off, its minimum and maximum replicas, the share of an
// @Description instance's Concurrency it keeps busy (target, percent) and how long its load stays low before it
// @Description scales in; its replicas now; and, when on but unable to act, why it is paused - its calls, read
// @Description from its stored logs, cannot be read.
// @Tags    App settings
// @Produce json
// @Id      getAppAutoscale
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Success 200 {object} appsettingsdto.GetAppAutoscaleResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/autoscale [get]
func (h *Handler) GetAppAutoscale(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appsettingsdto.NewGetAppAutoscaleReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appSettingsUC.GetAppAutoscale(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// UpdateAppAutoscale Updates a function's autoscale
// @Summary Updates a function's autoscale
// @Description Updates a function's autoscale. Turning it on needs its stored logs, which it reads its calls
// @Description from; it brings the function within its bounds at once, a stopped function staying stopped.
// @Description Turning it off leaves the replicas as they are.
// @Tags    App settings
// @Produce json
// @Id      updateAppAutoscale
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body appsettingsdto.UpdateAppAutoscaleReq true "request data"
// @Success 200 {object} appsettingsdto.UpdateAppAutoscaleResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/autoscale [put]
func (h *Handler) UpdateAppAutoscale(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appsettingsdto.NewUpdateAppAutoscaleReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appSettingsUC.UpdateAppAutoscale(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
