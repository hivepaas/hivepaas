package appactionhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appactionuc/appactiondto"
)

// SetAppRunning Sets app running status
// @Summary Sets app running status
// @Description Sets app running status
// @Tags    app_actions
// @Produce json
// @Id      appActionSetRunning
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body appactiondto.SetAppRunningReq true "request data"
// @Success 200 {object} appactiondto.SetAppRunningResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/running-status [post]
func (h *Handler) SetAppRunning(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeExecute)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appactiondto.NewSetAppRunningReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
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
