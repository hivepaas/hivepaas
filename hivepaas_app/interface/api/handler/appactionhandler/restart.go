package appactionhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appactionuc/appactiondto"
)

// RestartApp Restarts an app
// @Summary Restarts an app
// @Description Restarts an app
// @Tags    app_actions
// @Produce json
// @Id      appActionRestart
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   body body appactiondto.RestartAppReq true "request data"
// @Success 200 {object} appactiondto.RestartAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/restart [post]
func (h *Handler) RestartApp(ctx *gin.Context) {
	auth, projectID, appID, err := h.GetAuth(ctx, base.ActionTypeExecute, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appactiondto.NewRestartAppReq()
	req.ProjectID = projectID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appActionUC.RestartApp(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
