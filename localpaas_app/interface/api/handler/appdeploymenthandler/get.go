package appdeploymenthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"

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

// ListAppDeployment Lists app deployments
// @Summary Lists app deployments
// @Description Lists app deployments
// @Tags    app_deployments
// @Produce json
// @Id      listAppDeployment
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   status query string false "`status=<target>`"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} appdeploymentdto.ListDeploymentResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/deployments [get]
func (h *Handler) ListAppDeployment(ctx *gin.Context) {
	auth, projectID, appID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdeploymentdto.NewListDeploymentReq()
	req.ProjectID = projectID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appDeploymentUC.ListDeployment(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetAppDeploymentLogs Stream app deployment logs via websocket
// @Summary Stream app deployment logs via websocket
// @Description Stream deployment app logs via websocket
// @Tags    app_deployments
// @Produce json
// @Id      getAppDeploymentLogs
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   deploymentID path string true "deployment ID"
// @Param   follow query string false "`follow=true/false`"
// @Param   since query string false "`since=YYYY-MM-DDTHH:mm:SSZ`"
// @Param   duration query string false "`duration=24h` logs within the period"
// @Param   tail query int false "`tail=1000` to get last 1000 lines of logs"
// @Success 200 {object} appdeploymentdto.GetDeploymentLogsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/deployments/{deploymentID}/logs [get]
func (h *Handler) GetAppDeploymentLogs(ctx *gin.Context, mel *melody.Melody) {
	auth, projectID, appID, deploymentID, err := h.GetAuthForItem(ctx, base.ActionTypeRead, "deploymentID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdeploymentdto.NewGetDeploymentLogsReq()
	req.ProjectID = projectID
	req.AppID = appID
	req.DeploymentID = deploymentID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	isWebsocketReq := h.IsWebsocketRequest(ctx)
	if !isWebsocketReq {
		req.Follow = false // Not a websocket request, we don't support `follow` flag
	}

	resp, err := h.appDeploymentUC.GetDeploymentLogs(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	if !isWebsocketReq {
		// Not a websocket request, return data via body
		ctx.JSON(http.StatusOK, resp)
	} else {
		h.StreamAppLogs(ctx, resp.Data.Logs, resp.Data.LogChan, resp.Data.LogChanCloser, mel)
	}
}
