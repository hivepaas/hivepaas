package appdeploymenthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

// GetAppDeploymentLogsToken Gets a token for getting logs via websocket
// @Summary Gets a token for getting logs via websocket
// @Description Gets a token for getting logs via websocket
// @Tags    app_deployments
// @Produce json
// @Id      getAppDeploymentLogsToken
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   deploymentID path string true "deployment ID"
// @Success 200 {object} appdeploymentdto.GetDeploymentLogsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/deployments/{deploymentID}/logs/token [get]
func (h *Handler) GetAppDeploymentLogsToken(ctx *gin.Context) {
	auth, projectID, appID, deploymentID, err := h.GetAuthForItem(ctx, base.ActionTypeRead, "deploymentID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdeploymentdto.NewGetDeploymentLogsTokenReq()
	req.ProjectID = projectID
	req.AppID = appID
	req.DeploymentID = deploymentID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appDeploymentUC.GetDeploymentLogsToken(h.RequestCtx(ctx), auth, req)
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
// @Param   token query string false "`token=console-token`"
// @Param   follow query string false "`follow=true/false`"
// @Param   since query string false "`since=YYYY-MM-DDTHH:mm:SSZ`"
// @Param   duration query string false "`duration=24h` logs within the period"
// @Param   tail query int false "`tail=1000` to get last 1000 lines of logs"
// @Success 200 {object} appdeploymentdto.GetDeploymentLogsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/deployments/{deploymentID}/logs [get]
func (h *Handler) GetAppDeploymentLogs(ctx *gin.Context, mel *melody.Melody) {
	var auth *basedto.Auth
	var req *appdeploymentdto.GetDeploymentLogsReq
	var err error

	if ctx.Query("token") != "" {
		auth, req, err = h.parseAppDeploymentLogsReqWithToken(ctx)
	} else {
		auth, req, err = h.parseAppDeploymentLogsReqWithAuthHeader(ctx)
	}
	if err != nil {
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
		h.StreamAppLogs(ctx, resp.Data.StaticLogs, resp.Data.RealtimeLogsStream, resp.Data.LogsStreamCloser, mel)
	}
}

func (h *Handler) parseAppDeploymentLogsReqWithAuthHeader(
	ctx *gin.Context,
) (*basedto.Auth, *appdeploymentdto.GetDeploymentLogsReq, error) {
	auth, projectID, appID, deploymentID, err := h.GetAuthForItem(ctx, base.ActionTypeRead, "deploymentID")
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	req := appdeploymentdto.NewGetDeploymentLogsReq()
	req.ProjectID = projectID
	req.AppID = appID
	req.DeploymentID = deploymentID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	return auth, req, nil
}

//nolint:unparam
func (h *Handler) parseAppDeploymentLogsReqWithToken(
	ctx *gin.Context,
) (*basedto.Auth, *appdeploymentdto.GetDeploymentLogsReq, error) {
	projectID, err := h.ParseStringParam(ctx, "projectID")
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}
	appID, err := h.ParseStringParam(ctx, "appID")
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}
	deploymentID, err := h.ParseStringParam(ctx, "deploymentID")
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	req := appdeploymentdto.NewGetDeploymentLogsReq()
	req.ProjectID = projectID
	req.AppID = appID
	req.DeploymentID = deploymentID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	return nil, req, nil
}
