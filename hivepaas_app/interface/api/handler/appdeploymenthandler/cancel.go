package appdeploymenthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

// CancelAppDeployment Cancels app deployment
// @Summary Cancels app deployment
// @Description Cancels app deployment
// @Tags    app_deployments
// @Produce json
// @Id      cancelAppDeployment
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   deploymentID path string true "deployment ID"
// @Param   body body appdeploymentdto.CancelDeploymentReq true "request data"
// @Success 200 {object} appdeploymentdto.CancelDeploymentResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/deployments/{deploymentID}/cancel [post]
func (h *Handler) CancelAppDeployment(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, itemID, err := h.GetAuthForItem(ctx, base.ActionTypeWrite, "deploymentID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdeploymentdto.NewCancelDeploymentReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	req.DeploymentID = itemID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appDeploymentUC.CancelDeployment(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
