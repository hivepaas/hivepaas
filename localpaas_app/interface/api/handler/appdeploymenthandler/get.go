package appdeploymenthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

// GetAppDeployment Gets app deployment
// @Summary Gets app deployment
// @Description Gets app deployment
// @Tags    app_deployments
// @Produce json
// @Id      getAppDeployment
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   deploymentID path string true "deployment ID"
// @Success 200 {object} appdeploymentdto.GetDeploymentResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/deployments/{deploymentID} [get]
func (h *Handler) GetAppDeployment(ctx *gin.Context) {
	auth, projectID, appID, deploymentID, err := h.GetAuthForItem(ctx, base.ActionTypeRead, "deploymentID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdeploymentdto.NewGetDeploymentReq()
	req.ProjectID = projectID
	req.AppID = appID
	req.DeploymentID = deploymentID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appDeploymentUC.GetDeployment(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetAppDeploymentStatus Gets app deployment status
// @Summary Gets app deployment status
// @Description Gets app deployment status
// @Tags    app_deployments
// @Produce json
// @Produce plain
// @Id      getAppDeploymentStatus
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   deploymentID path string true "deployment ID"
// @Success 200 {object} appdeploymentdto.GetDeploymentStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/deployments/{deploymentID}/status [get]
func (h *Handler) GetAppDeploymentStatus(ctx *gin.Context) {
	auth, projectID, appID, deploymentID, err := h.GetAuthForItem(ctx, base.ActionTypeRead, "deploymentID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdeploymentdto.NewGetDeploymentStatusReq()
	req.ProjectID = projectID
	req.AppID = appID
	req.DeploymentID = deploymentID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appDeploymentUC.GetDeploymentStatus(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	if ctx.ContentType() == "text/plain" {
		ctx.String(http.StatusOK, "status=%v", resp.Data.Status)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
