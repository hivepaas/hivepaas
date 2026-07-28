package apphandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// UpdateAppStatus Updates app status
// @Summary Updates app status
// @Description Updates app status
// @Tags    apps
// @Produce json
// @Id      updateAppStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body appdto.UpdateAppStatusReq true "request data"
// @Success 200 {object} appdto.UpdateAppStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/status [put]
func (h *Handler) UpdateAppStatus(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewUpdateAppStatusReq()
	req.AppID = appID
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appUC.UpdateAppStatus(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
