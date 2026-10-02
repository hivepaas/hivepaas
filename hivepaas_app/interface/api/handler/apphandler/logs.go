package apphandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// GetAppLogsInfo Gets log info
// @Summary Gets log info
// @Description Gets log info
// @Tags    Apps
// @Produce json
// @Id      getAppLogsInfo
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Success 200 {object} appdto.GetAppLogsInfoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/logs/info [get]
func (h *Handler) GetAppLogsInfo(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewGetAppLogsInfoReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
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

// GetAppLogs Stream app logs via websocket
// @Summary Stream app logs via websocket
// @Description Stream app logs via websocket.
// @Description Every read is bounded: `tail` defaults to 1000 lines and may not exceed 5000.
// @Description Reading further back than one response holds is the history endpoint, which pages.
// @Tags    Apps
// @Produce json
// @Id      getAppLogs
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   taskId query string false "`taskId=<task-id>`"
// @Param   follow query string false "`follow=true/false`"
// @Param   since query string false "`since=YYYY-MM-DDTHH:mm:SSZ`"
// @Param   duration query int false "`duration=` logs within the period"
// @Param   tail query int false "`tail=1000` for the last 1000 lines, 1-5000, default 1000"
// @Success 200 {object} appdto.GetAppLogsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/logs [get]
func (h *Handler) GetAppLogs(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewGetAppLogsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
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
		h.StreamAppLogs(ctx, resp.Data.StaticLogs, resp.Data.LogsStream, resp.Data.LogsStreamCloser)
	}
}

// GetAppLogHistory Gets stored app logs
// @Summary Gets stored app logs
// @Description Searches the logs collected for the app, including those of containers that no longer exist.
// @Description Parameters are structured; no query text is accepted.
// @Description `search` is matched from the start of a token, which the backend answers from its index.
// @Description `regex=true` reads it as a regular expression instead: slower, because it is read row by
// @Description row, and a malformed expression comes back as an invalid query, not as an empty result.
// @Tags    Apps
// @Produce json
// @Id      getAppLogHistory
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   start query string false "`start=YYYY-MM-DDTHH:mm:SSZ`, default one hour before end"
// @Param   end query string false "`end=YYYY-MM-DDTHH:mm:SSZ`, default now; pass `nextEnd` to page back"
// @Param   limit query int false "lines to return, 1-5000, default 500"
// @Param   search query string false "text to match in the message"
// @Param   regex query bool false "`regex=true` reads search as a regular expression"
// @Param   matchCase query bool false "`matchCase=true` compares case-sensitively"
// @Param   levels query string false "comma-separated: trace,debug,info,warn,warning,error,fatal,panic"
// @Param   streams query string false "comma-separated: stdout,stderr"
// @Success 200 {object} appdto.GetAppLogHistoryResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/logs/history [get]
func (h *Handler) GetAppLogHistory(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewGetAppLogHistoryReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appUC.GetAppLogHistory(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetAppHTTPMetrics Gets an app's HTTP metrics
// @Summary Gets an app's HTTP metrics
// @Description Counts an app's requests over a range ending now, from the proxy's access log: how many,
// @Description how many the client got a 4xx and a 5xx for, how many the proxy could not get to the app at
// @Description all (unreachable), and how long they took end to end - p50, p95 and p99, in milliseconds,
// @Description close rather than exact. In totals, by method and path (numbers and ids replaced by :n and
// @Description :id), by replica, and as series: one point per step, oldest first, as for a function's
// @Description metrics. `available` is false, with a `reason`, when they cannot be counted: the app has no
// @Description domain (not-exposed), the proxy's access log is off, not JSON or unlabelled, or the logs
// @Description cannot be read.
// @Tags    Apps
// @Produce json
// @Id      getAppHTTPMetrics
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   range query string false "1h, 6h, 24h or 7d; default 24h"
// @Success 200 {object} appdto.GetAppHTTPMetricsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/http-metrics [get]
func (h *Handler) GetAppHTTPMetrics(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewGetAppHTTPMetricsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appUC.GetAppHTTPMetrics(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetFunctionMetrics Gets a function's metrics
// @Summary Gets a function's metrics
// @Description Counts a function's calls over a range ending now, from the invocation line its runtime writes
// @Description for every call into its logs: how many, how many failed - an outcome other than ok - how many
// @Description the handler answered 5xx, and the handler's duration's p50, p95 and p99, in milliseconds, which
// @Description are close rather than exact. One point per step, oldest first: 1 min for 1h, 5 min for 6h,
// @Description 15 min for 24h, 1 h for 7d. When the logs cannot be read, `available` is false and `reason`
// @Description says why, as for the logs' history.
// @Tags    Apps
// @Produce json
// @Id      getFunctionMetrics
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   range query string false "1h, 6h, 24h or 7d; default 24h"
// @Success 200 {object} appdto.GetFunctionMetricsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/function-metrics [get]
func (h *Handler) GetFunctionMetrics(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewGetFunctionMetricsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appUC.GetFunctionMetrics(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
