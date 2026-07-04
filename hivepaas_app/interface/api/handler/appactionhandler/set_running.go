package appactionhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appactionuc/appactiondto"
)

// SetAppRunning Sets app running status
// @Summary Sets app running status
// @Description Sets app running status
// @Tags    app_actions
// @Produce json
// @Id      appActionSetRunning
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   body body appactiondto.SetAppRunningReq true "request data"
// @Success 200 {object} appactiondto.SetAppRunningResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/running-status [post]
func (h *Handler) SetAppRunning(ctx *gin.Context) {
	auth, projectID, appID, err := h.GetAuth(ctx, base.ActionTypeExecute, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appactiondto.NewSetAppRunningReq()
	req.ProjectID = projectID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appActionUC.SetAppRunning(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
