package appdeploymenthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

// GetAppDeployment Gets app deployment
// @Summary Gets app deployment
// @Description Gets app deployment
// @Tags    App deployments
// @Produce json
// @Id      getAppDeployment
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   deploymentID path string true "deployment ID"
// @Success 200 {object} appdeploymentdto.GetDeploymentResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/deployments/{deploymentID} [get]
func (h *Handler) GetAppDeployment(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, deploymentID, err := h.GetAuthForItem(ctx, base.ActionTypeRead, "deploymentID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdeploymentdto.NewGetDeploymentReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
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
// @Tags    App deployments
// @Produce json
// @Produce plain
// @Id      getAppDeploymentStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   deploymentID path string true "deployment ID"
// @Success 200 {object} appdeploymentdto.GetDeploymentStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/deployments/{deploymentID}/status [get]
func (h *Handler) GetAppDeploymentStatus(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, deploymentID, err := h.GetAuthForItem(ctx, base.ActionTypeRead, "deploymentID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdeploymentdto.NewGetDeploymentStatusReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
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

// GetActiveAppDeployment Gets the app's deployment that has not ended
// @Summary Gets the app's deployment that has not ended
// @Description The deployment running - in-progress - or else the next to run, not-started; data is null when none
// @Description is queued or running. Light enough to be asked every few seconds.
// @Tags    App deployments
// @Produce json
// @Id      getActiveAppDeployment
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Success 200 {object} appdeploymentdto.GetActiveDeploymentResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/deployments/active [get]
func (h *Handler) GetActiveAppDeployment(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdeploymentdto.NewGetActiveDeploymentReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appDeploymentUC.GetActiveDeployment(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
