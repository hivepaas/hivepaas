package apphandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

// GetAppLogsInfo Gets log info
// @Summary Gets log info
// @Description Gets log info
// @Tags    apps
// @Produce json
// @Id      getAppLogsInfo
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Success 200 {object} appdto.GetAppLogsInfoResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/logs/info [get]
func (h *Handler) GetAppLogsInfo(ctx *gin.Context) {
	auth, projectID, appID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewGetAppLogsInfoReq()
	req.ProjectID = projectID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appUC.GetAppLogsInfo(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetAppLogsToken Gets a token for getting logs via websocket
// @Summary Gets a token for getting logs via websocket
// @Description Gets a token for getting logs via websocket
// @Tags    apps
// @Produce json
// @Id      getAppLogsToken
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Success 200 {object} appdto.GetAppLogsTokenResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/logs/token [get]
func (h *Handler) GetAppLogsToken(ctx *gin.Context) {
	auth, projectID, appID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewGetAppLogsTokenReq()
	req.ProjectID = projectID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appUC.GetAppLogsToken(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetAppLogs Stream app logs via websocket
// @Summary Stream app logs via websocket
// @Description Stream app logs via websocket
// @Tags    apps
// @Produce json
// @Id      getAppLogs
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   taskId query string false "`taskId=<task-id>`"
// @Param   follow query string false "`follow=true/false`"
// @Param   since query string false "`since=YYYY-MM-DDTHH:mm:SSZ`"
// @Param   duration query int false "`duration=` logs within the period"
// @Param   tail query int false "`tail=1000` to get last 1000 lines of logs"
// @Success 200 {object} appdto.GetAppLogsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/logs [get]
func (h *Handler) GetAppLogs(ctx *gin.Context, mel *melody.Melody) {
	var auth *basedto.Auth
	var req *appdto.GetAppLogsReq
	var err error

	if ctx.Query("token") != "" {
		auth, req, err = h.parseAppLogsReqWithToken(ctx)
	} else {
		auth, req, err = h.parseAppLogsReqWithAuthHeader(ctx)
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	isWebsocketReq := h.IsWebsocketRequest(ctx)
	if !isWebsocketReq {
		req.Follow = false // Not a websocket request, we don't support `follow` flag
	}

	resp, err := h.appUC.GetAppLogs(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	if !isWebsocketReq {
		// Not a websocket request, return data via body
		ctx.JSON(http.StatusOK, resp)
	} else {
		h.StreamAppLogs(ctx, resp.Data.StaticLogs, resp.Data.LogsStream, resp.Data.LogsStreamCloser, mel)
	}
}

func (h *Handler) parseAppLogsReqWithAuthHeader(
	ctx *gin.Context,
) (*basedto.Auth, *appdto.GetAppLogsReq, error) {
	auth, projectID, appID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	req := appdto.NewGetAppLogsReq()
	req.ProjectID = projectID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	return auth, req, nil
}

//nolint:unparam
func (h *Handler) parseAppLogsReqWithToken(
	ctx *gin.Context,
) (*basedto.Auth, *appdto.GetAppLogsReq, error) {
	projectID, err := h.ParseStringParam(ctx, "projectID")
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}
	appID, err := h.ParseStringParam(ctx, "appID")
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	req := appdto.NewGetAppLogsReq()
	req.ProjectID = projectID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	return nil, req, nil
}
