package appdeploymenthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

// DeployApp Deploys an app
// @Summary Deploys an app
// @Description Deploys an app
// @Tags    app_deployments
// @Produce json
// @Id      apiDeployApp
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   body body appdeploymentdto.DeployAppReq true "request data"
// @Success 200 {object} appdeploymentdto.DeployAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/deploy [post]
func (h *Handler) DeployApp(ctx *gin.Context) {
	auth, projectID, appID, err := h.GetAuth(ctx, base.ActionTypeExecute, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdeploymentdto.NewDeployAppReq()
	req.ProjectID = projectID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appDeploymentUC.DeployApp(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
